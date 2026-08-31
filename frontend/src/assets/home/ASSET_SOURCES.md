# 首页视觉素材来源

## 首页背景图（无静态水波）

- `hero-rain-celadon-light.png`、`hero-night-rain-dark.png`：基于原有雨幕背景做局部图像编辑，仅移除背景中预置的同心水波、溅水点与水滴柱，保留原有色调、雨丝和水面层次。
- 页面用途：背景保持静止且无预置水波；水波只由雨滴/花瓣落水和鼠标点击事件在运行时触发。

## `source-pink-flower-cc0.webp`

- 来源页：<https://purepng.com/photo/30004/nature-flower>
- 原始文件：<https://purepng.com/public/uploads/large/flower-4ms.png>
- 授权：CC0；来源页标注可自由使用、允许商业用途且无需署名。
- 优化方式：使用 [images.weserv.nl](https://images.weserv.nl/) 将原始 `1644×1562` PNG 等比缩放为 `720×684` 透明 WebP；文件由约 `2.08 MB` 降至约 `41 KB`，没有生成或重绘内容。
- 页面用途：从透明底真实花卉照片的外层花瓣区域取景，并按单片花瓣轮廓遮罩，在运行时形成不同角度、大小和明暗的飘落花瓣。未使用矢量花瓣或 AI 生成花瓣替代。

## 点击水波运动参考

- [Circular waves.jpg](https://commons.wikimedia.org/wiki/File:Circular_waves.jpg)：真实水滴落水照片，作者 Saral Shots，CC0。
- [Drop creates waves.jpg](https://commons.wikimedia.org/wiki/File:Drop_creates_waves.jpg)：真实水滴与同心波照片，作者 Sindugab，CC0。
- 页面用途：两张实拍资料只用于校准运动阶段与衰减节奏，不直接嵌入页面。点击效果采用“冲击环 → 主波 → 尾波”三层现有透明水纹栅格叠加，约 `4.5s` 完成扩散与消散；未使用 AI 生成参考图。
