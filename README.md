# MyDebugger

一个调用 delve 的接口，加了 TUI 的水项目



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

- q/quit：退出调试器
- b/break：添加断点，支持地址断点和函数名断点，也支持给断点命名
- c/continue：运行至下一个断点出处
- si/step-instruction：汇编层面的单步执行
- n/next：单步执行，但不进入函数内
- so/step-out：跳出当前函数，即执行完当前函数，并返回上一调用栈
- clear：清除断点，支持地址、id、断点名
- clear-all：清除所有断点
- x：查看内存，格式化输出，例如：`x gx 0xc00007df78` ，类似 gdb 的结果
- "": 继续上一步的指令
- d/disassembly：反汇编地址处的值，显示在反汇编窗口上







