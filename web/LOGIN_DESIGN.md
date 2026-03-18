# 登录页设计说明

## 配色方案

### 暗黑模式（默认）
- **背景**: `#0a0e1a` 深邃科技蓝黑
- **玻璃卡片**: `rgba(17, 24, 39, 0.7)` 半透明深灰 + 20px 模糊
- **主题色渐变**: 青绿 `#10b981` → 橙金 `#f59e0b`
- **边框微光**: `rgba(16, 185, 129, 0.2)` 青绿半透明
- **文字**: 主文字 `#f9fafb` / 次要文字 `#9ca3af`

### 亮色模式
- **背景**: `#f0f9ff` 天空蓝渐变
- **玻璃卡片**: `rgba(255, 255, 255, 0.8)` 半透明白 + 20px 模糊
- **主题色渐变**: 蓝色 `#3b82f6` → 橙金 `#f59e0b`
- **边框微光**: `rgba(59, 130, 246, 0.2)` 蓝色半透明
- **文字**: 主文字 `#111827` / 次要文字 `#6b7280`

## 动画细节

### 1. 粒子背景
- 50个流体粒子随机运动
- 距离<120px自动连线形成网络
- 暗黑模式青绿色 / 亮色模式蓝色
- 自适应窗口大小

### 2. 输入框焦点动画
```css
/* 焦点时 */
- 边框变为主题色
- 外发光 4px 青绿/蓝色光晕
- 平滑缩放 scale(1.02)
- 过渡时间 0.3s ease
```

### 3. 错误提示动画
```css
/* 错误时 */
- 输入框抖动 shake 0.4s
- 错误文字淡入 fadeIn 0.3s
- 边框变红色
```

### 4. 按钮交互
```css
/* 悬停时 */
- 上浮 translateY(-2px)
- 渐变反转（通过 ::before 伪元素）
- 外发光 8px 主题色光晕
- 加载时显示旋转 spinner
```

### 5. Logo 浮动
```css
/* 持续动画 */
- 上下浮动 3s 无限循环
- 幅度 10px
- ease-in-out 缓动
```

### 6. 主题切换按钮
```css
/* 悬停时 */
- 放大 scale(1.1)
- 旋转 15deg
- 外发光主题色
```

## 布局结构

```
login-container (全屏容器)
├── canvas (粒子背景)
├── theme-toggle (右上角主题切换)
└── login-card (玻璃态卡片)
    ├── card-left (左侧品牌区)
    │   ├── logo-icon (SVG Logo)
    │   ├── brand-title (MrChat)
    │   └── brand-tagline (聚合智能·对话未来)
    └── card-right (右侧表单区)
        ├── form-title (欢迎回来)
        ├── login-form
        │   ├── form-group (邮箱/用户名)
        │   ├── form-group (密码)
        │   ├── error-message (错误提示)
        │   └── submit-btn (登录按钮)
        └── signup-link (注册链接)
```

## 响应式设计

- 桌面端: 左右分栏布局 (1000px max-width)
- 移动端 (<768px): 上下堆叠布局
- 卡片宽度自适应 90%

## 使用方式

主题会自动保存到 localStorage，刷新后保持用户选择。

```typescript
import { useTheme } from '@/composables/useTheme'

const { theme, toggleTheme, isDark } = useTheme()
```
