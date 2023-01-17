# MyDebugger

一个调用 delve 的接口，加了 TUI 的水项目



## 2023/01/17 更新

1. 抽出 `Monitors.go` 用于监控，且 `monitor` 的 size 默认为 4
2. 添加功能：当监控的数据发生变化的时候，自动显示 监视器 界面，并在界面上标红
3. 调整了 `initCommand` 函数的初始化布局，使用**内存逃逸**。
4. 把 `GetDataFromAddress` 放到 `api.go` 中，因为感觉更像是 api 提供的能力
5. 添加监控的时候可以引用寄存器的值和运算，例如 `monitor $Rsp+0x20`

![image-20230117163244827](https://s2.loli.net/2023/01/17/GEHAbzIXJuRxNUF.png)



## 2023/01/16 更新

1. 添加 `print <address> <size>` 命令，用于打印某个地址处的值，以 `size` 的方式

![image-20230116102331197](https://s2.loli.net/2023/01/16/u5QqO4rIWR6VJEB.png)

2. 添加 `monitor <address> <size>` 命令，用于监视某个地址处的值，发生变化时，会变红（todo：用协程实时监控，发生变化就提示）

![image-20230116102442897](https://s2.loli.net/2023/01/16/8gGdCMTjkUm4wt3.png)

3. 拆出 `UICommand.go` ，仅用于处理命令的函数
4. 拆出 `UIView.go` ，仅用于处理UI视图函数
5. Error 用 `channel` 的方式传递，接收到 error 之后，显示在右下角





## 2023/01/12 更新

1. 添加 `focus <id>` 命令，让用户可以切换视图，当前窗体由红色边框标注，按下 `enter` 或者 `esc` 即返回命令输入窗口，例如在 反汇编窗口可以上下键移动（但不是实时，是缓存数据，这里有待改进）

![image-20230112113326345](https://s2.loli.net/2023/01/12/frpsoKOLNPHxE4u.png)

2. 更改 focus 提供两种方法：
   1. 重新创建 `ui.grid` 对象
   2. 使用反射更改 item 的 focus 值



## 2023/01/11 更新

1. 将 ViewInfo.go 拆出
2. 改变输入命令后的 `dealWithEnter` 函数对命令的查找逻辑，从 `switch` 变更到 `map[string]*Command` ，利用回调函数执行。并将所有的 `switch` 单独拆成函数
3. 命令提示字典从 map 中自动获取
4. 执行指令出现 error 时，显示在右下角的窗体中，例如输入 `clear 4` 但是没有 ID 为 4 的断点时：

![image-20230111110511590](https://s2.loli.net/2023/01/11/ihPr1OQk4tevLHZ.png)

5. 添加 help 信息，执行 `help <command>`  可以查看对应命令的用法，例如执行 help b

![image-20230111110624079](https://s2.loli.net/2023/01/11/PqpTE6cYWId5SxU.png)



## 2023/01/10 更新

1. 将 Data.go 整合至 UI.go 里，UI.go 里将 view 与 data 绑定，为了到时候更换视图方便，比如说现在放内存的窗口，想用来监视变量。
2. 添加历史命令界面
3. 添加断点界面

## 2023/01/09 更新

1. 标注指令所在的函数的名称，无法获取到的不标注，以黄色注释标注

![image-20230109160952287](https://s2.loli.net/2023/01/09/HQo79TC3PS4LvEY.png)



2. 在执行命令之后，寄存器的值如果发生了变化，则会用红色标注（Rip寄存器除外）

​	例如执行了 `sub rsp, 0xb8` 之后，寄存器 Rsp 的值发生变化，标注为红色

![image-20230109160801158](https://s2.loli.net/2023/01/09/MsSLNB57qA63rKU.png)



3. api.go/ui.go/data.go 使用 channel

api.go里 `Continue` 函数使用 channel 代替 Sleep，解决运行时间和效率问题

ui.go 和 data.go 在获取数据和更新界面的时候使用 `errgroup` 来处理异常（纯学习，感觉用不到）

4. 添加反汇编 地址 指令，详细见 d/disassembly

## 使用截图

右下角还没想好放什么，先空着

![image-20230107104624140](https://s2.loli.net/2023/01/07/XvjAzFHs2t5xouK.png)

### 反汇编

最左边是指令地址，中间的 `#` 表示是否在该地址处有断点，右边是指令

当前 Rip 所指向的命令由<font color='red'>红色</font>标注

### 寄存器

通过接口获取到的寄存器有很多个，在这里就截取到 `R15` 寄存器，其他的感觉使用场景有限

### 内存

默认是看 `Rsp` 寄存器内值的地址，也就是栈上的数据

#### 输入命令

详情请见 `UI.go/initCommands` 函数







