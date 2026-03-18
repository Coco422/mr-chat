import type { Router } from 'vue-router'

type PerfMetricKind = 'navigation' | 'paint' | 'web-vital' | 'route' | 'api'

export interface PerfMetric {
  name: string
  value: number
  unit: 'ms' | 'score'
  kind: PerfMetricKind
  timestamp: number
  extra?: Record<string, unknown>
}

export interface PerformanceMonitorOptions {
  enabled?: boolean
  debug?: boolean
  router?: Router
  onMetric?: (metric: PerfMetric) => void
}

// 模块级 reporter。保持全局单例，方便在不同文件里复用，不需要复杂依赖注入。
let globalReporter: ((metric: PerfMetric) => void) | null = null

/**
 * 初始化前端性能埋点。
 *
 * 设计原则：
 * 1) 代码尽量少，避免理解成本高。
 * 2) 默认只打印到控制台，先能看见数据，再决定是否上报后端。
 * 3) 每个指标都拆成单独函数，后续删改任何一个都不会影响其他逻辑。
 */
export function setupPerformanceMonitor(options: PerformanceMonitorOptions = {}) {
  const enabled = options.enabled ?? true
  if (!enabled || typeof window === 'undefined' || typeof performance === 'undefined') {
    return
  }

  globalReporter = (metric) => {
    if (options.debug) {
      if (metric.kind === 'api') {
        logApiMetric(metric)
      } else {
        console.info('[perf]', metric)
      }
    }
    options.onMetric?.(metric)
  }

  reportNavigationMetrics()
  observePaintMetrics()
  observeLCP()
  observeCLS()
  observeFID()

  if (options.router) {
    observeRouteNavigation(options.router)
  }
}

/**
 * 对外暴露一个“手动上报”方法：
 * - 你未来如果想埋点按钮点击耗时、接口耗时，可以直接调用这个函数。
 * - 不想用也没关系，不会影响基础指标采集。
 */
export function reportPerfMetric(metric: Omit<PerfMetric, 'timestamp'>) {
  if (!globalReporter) {
    return
  }

  globalReporter({
    ...metric,
    timestamp: Date.now()
  })
}

/**
 * Navigation Timing（导航性能）
 * 这里拿的是“页面首屏加载”的关键阶段：
 * - ttfb: 从发起请求到收到首字节（后端和网络慢会直接影响这里）
 * - dom_content_loaded: DOM 解析完成
 * - load: 页面资源加载完成
 */
function reportNavigationMetrics() {
  const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined
  if (!nav) {
    return
  }

  reportPerfMetric({
    name: 'ttfb',
    value: nav.responseStart,
    unit: 'ms',
    kind: 'navigation'
  })

  reportPerfMetric({
    name: 'dom_content_loaded',
    value: nav.domContentLoadedEventEnd,
    unit: 'ms',
    kind: 'navigation'
  })

  reportPerfMetric({
    name: 'window_load',
    value: nav.loadEventEnd,
    unit: 'ms',
    kind: 'navigation'
  })
}

/**
 * Paint Timing（绘制性能）
 * 常用的是 FCP：首次内容绘制时间。
 * 数值越小，用户越快看到“页面有内容了”。
 */
function observePaintMetrics() {
  safeObserve(
    'paint',
    (entry) => {
      if (entry.name !== 'first-contentful-paint') {
        return
      }

      reportPerfMetric({
        name: 'fcp',
        value: entry.startTime,
        unit: 'ms',
        kind: 'paint'
      })
    },
    { buffered: true }
  )
}

/**
 * LCP（Largest Contentful Paint）
 * 代表“首屏最大内容元素”何时渲染完成，通常是用户感知加载速度的核心指标之一。
 */
function observeLCP() {
  let lcpValue = 0

  safeObserve(
    'largest-contentful-paint',
    (entry) => {
      lcpValue = entry.startTime
    },
    { buffered: true }
  )

  // 页面即将隐藏时再上报最终值，避免中途多次更新造成噪音。
  onPageHidden(() => {
    if (lcpValue <= 0) {
      return
    }
    reportPerfMetric({
      name: 'lcp',
      value: lcpValue,
      unit: 'ms',
      kind: 'web-vital'
    })
  })
}

/**
 * CLS（Cumulative Layout Shift）
 * 代表页面“抖动”程度；这个值越低越稳定。
 * 这里过滤了用户主动交互触发的位移（hadRecentInput），避免误报。
 */
function observeCLS() {
  let clsValue = 0

  safeObserve(
    'layout-shift',
    (entry) => {
      const layoutShiftEntry = entry as PerformanceEntry & {
        value?: number
        hadRecentInput?: boolean
      }
      if (layoutShiftEntry.hadRecentInput) {
        return
      }
      clsValue += layoutShiftEntry.value ?? 0
    },
    { buffered: true }
  )

  onPageHidden(() => {
    reportPerfMetric({
      name: 'cls',
      value: Number(clsValue.toFixed(4)),
      unit: 'score',
      kind: 'web-vital'
    })
  })
}

/**
 * FID（First Input Delay）
 * 用户首次交互（点击、输入）到浏览器真正开始处理事件的延迟。
 */
function observeFID() {
  safeObserve(
    'first-input',
    (entry) => {
      const firstInputEntry = entry as PerformanceEntry & {
        processingStart?: number
      }
      const processingStart = firstInputEntry.processingStart ?? entry.startTime
      reportPerfMetric({
        name: 'fid',
        value: processingStart - entry.startTime,
        unit: 'ms',
        kind: 'web-vital'
      })
    },
    { buffered: true }
  )
}

/**
 * 路由切换耗时（Vue 场景下最直观）
 * beforeEach 记录开始时间，afterEach 计算完成耗时。
 * 可以快速定位“哪个页面切换慢”。
 */
function observeRouteNavigation(router: Router) {
  let routeStartTime = 0

  router.beforeEach(() => {
    routeStartTime = performance.now()
    return true
  })

  router.afterEach((to, from, failure) => {
    if (routeStartTime <= 0) {
      return
    }

    reportPerfMetric({
      name: 'route_change',
      value: performance.now() - routeStartTime,
      unit: 'ms',
      kind: 'route',
      extra: {
        from: from.fullPath,
        to: to.fullPath,
        failed: Boolean(failure)
      }
    })

    routeStartTime = 0
  })
}

/**
 * 安全包装 PerformanceObserver：
 * - 某些浏览器不支持某些 entryType，直接 observe 会抛错。
 * - 统一在这里兜底，保证业务代码不需要处理兼容性细节。
 */
function safeObserve(
  entryType: string,
  onEntry: (entry: PerformanceEntry) => void,
  options: PerformanceObserverInit = {}
) {
  if (typeof PerformanceObserver === 'undefined') {
    return
  }

  const supportedTypes = PerformanceObserver.supportedEntryTypes ?? []
  if (!supportedTypes.includes(entryType)) {
    return
  }

  try {
    const observer = new PerformanceObserver((entryList) => {
      for (const entry of entryList.getEntries()) {
        onEntry(entry)
      }
    })

    observer.observe({
      type: entryType,
      buffered: options.buffered
    })
  } catch {
    // 静默失败，避免埋点影响业务。
  }
}

/**
 * 页面隐藏时回调（切标签、关闭页面等）。
 * web-vitals 通常在这个时机上报最终值，更稳定。
 */
function onPageHidden(callback: () => void) {
  const runOnce = () => {
    callback()
    document.removeEventListener('visibilitychange', onVisibilityChange, true)
    window.removeEventListener('pagehide', runOnce, true)
  }

  const onVisibilityChange = () => {
    if (document.visibilityState === 'hidden') {
      runOnce()
    }
  }

  document.addEventListener('visibilitychange', onVisibilityChange, true)
  window.addEventListener('pagehide', runOnce, true)
}

function logApiMetric(metric: PerfMetric) {
  const method = String(metric.extra?.method ?? 'GET')
  const path = String(metric.extra?.path ?? '')
  const status = String(metric.extra?.status ?? '-')
  const success = Boolean(metric.extra?.success)
  const label = success ? 'ok' : 'error'

  console.info(
    `[api] ${label} ${method} ${path} status=${status} duration=${metric.value.toFixed(1)}ms`,
    metric
  )
}
