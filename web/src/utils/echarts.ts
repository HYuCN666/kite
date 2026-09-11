// 引入 echarts 核心模块
import * as echarts from 'echarts/core'

// 引入折线图
import { LineChart } from 'echarts/charts'

// 引入提示框、标题、直角坐标系、图例组件
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
} from 'echarts/components'

// 引入 Canvas 渲染器
import { CanvasRenderer } from 'echarts/renderers'

// 注册必须的组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  LineChart,
  CanvasRenderer
])

export default echarts
