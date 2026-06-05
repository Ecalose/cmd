package cmd

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gospider007/conf"
	"github.com/gospider007/gson"
	"github.com/gospider007/tools"
)

type ClientOption struct {
	Name          string        //程序执行的名字
	Args          []string      //程序的执行参数
	Dir           string        //程序执行的位置
	TimeOut       time.Duration //程序超时时间
	CloseCallBack func()        //关闭时执行的函数
	WatchDog      bool          //是否开启看狗模式
}
type Client struct {
	err           error
	cmd           *exec.Cmd
	ctx           context.Context
	cnl           context.CancelFunc
	closeCallBack func() //关闭时执行的函数
}

//go:embed cmd/bin/watchdog-linux-amd64
var watchdogLinux []byte

//go:embed cmd/bin/watchdog-mac-arm64
var watchdogMac []byte

//go:embed cmd/bin/watchdog-windows-amd64.exe
var watchdogWin []byte
var watchDogVersion = "1.0.1"

// 普通的cmd 客户端
func NewClient(pre_ctx context.Context, option ClientOption) (*Client, error) {
	if option.WatchDog {
		var watchName string
		if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
			watchName = "linux-amd64"
		} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
			watchName = "mac-arm64"
		} else if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
			watchName = "windows-amd64.exe"
		} else {
			return nil, errors.New("watchdog not support this os")
		}
		watchName = fmt.Sprintf(".watchdog-%s-%s", watchDogVersion, watchName)
		userDir, err := conf.GetMainDirPath()
		if err != nil {
			return nil, err
		}
		filePath := tools.PathJoin(userDir, watchName)
		if !tools.PathExist(filePath) {
			switch runtime.GOOS {
			case "linux":
				err = os.WriteFile(filePath, watchdogLinux, 0777)
			case "darwin":
				err = os.WriteFile(filePath, watchdogMac, 0777)
			case "windows":
				err = os.WriteFile(filePath, watchdogWin, 0777)
			}
			if err != nil {
				return nil, err
			}
		}
		args := []string{strconv.Itoa(os.Getpid()), option.Name}
		args = append(args, option.Args...)
		option.Name = filePath
		option.Args = args
		option.WatchDog = false
		return NewClient(pre_ctx, option)
	}
	var ctx context.Context
	var cnl context.CancelFunc
	var err error
	if pre_ctx == nil {
		pre_ctx = context.TODO()
	}

	if option.TimeOut != 0 {
		ctx, cnl = context.WithTimeout(pre_ctx, option.TimeOut)
	} else {
		ctx, cnl = context.WithCancel(pre_ctx)
	}
	cmd := exec.CommandContext(ctx, option.Name, option.Args...)
	setAttr(cmd)
	cmd.Env = os.Environ()
	cmd.Dir = option.Dir
	result := &Client{
		cmd:           cmd,
		ctx:           ctx,
		cnl:           cnl,
		closeCallBack: option.CloseCallBack,
	}
	go tools.Signal(ctx, result.Close)
	return result, err
}

var ErrClosed = errors.New("client closed")

//go:embed cmdPipJsScript.js
var cmdPipJsScript []byte

//go:embed cmdPipPyScript.py
var cmdPipPyScript []byte

var scriptVersion = "025"

type JyClient struct {
	client *Client
	write  io.WriteCloser
	read   *bufio.Reader
	lock   sync.Mutex
}
type PyClientOption struct {
	PythonPath string   //python 的路径,ex: c:/python.exe
	ModulePath []string //python包搜索路径,如果出现搜索不到包的情况,手动在这里加入路径哈
}

// 创建py解析器
func NewPyClient(pre_ctx context.Context, options ...PyClientOption) (*JyClient, error) {
	var option PyClientOption
	if len(options) > 0 {
		option = options[0]
	}
	if option.PythonPath == "" {
		option.PythonPath = "python3"
	}
	nowDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if option.ModulePath == nil {
		option.ModulePath = []string{nowDir}
	} else {
		option.ModulePath = append(option.ModulePath, nowDir)
	}
	userDir, err := conf.GetMainDirPath()
	if err != nil {
		return nil, err
	}
	filePath := tools.PathJoin(userDir, fmt.Sprintf(".cmdPipPyScript%s.py", scriptVersion))
	if !tools.PathExist(filePath) {
		err := os.WriteFile(filePath, cmdPipPyScript, 0777)
		if err != nil {
			return nil, err
		}
	}
	cli, err := NewClient(pre_ctx, ClientOption{
		Name: option.PythonPath,
		Args: []string{"-u", filePath},
	})

	if err != nil {
		return nil, err
	}
	writeBody, err := cli.StdInPipe()
	if err != nil {
		return nil, err
	}
	readPip, err := cli.StdOutPipe()
	if err != nil {
		return nil, err
	}
	go cli.Run()
	pyCli := &JyClient{
		client: cli,
		read:   bufio.NewReader(readPip),
		write:  writeBody,
	}
	return pyCli, pyCli.init(pre_ctx, option.ModulePath)
}

type JsClientOption struct {
	NodePath   string   //node 的路径,ex: c:/node.exe
	ModulePath []string //node包搜索路径,如果出现搜索不到包的情况,手动在这里加入路径哈
}
type writeCloser struct {
	io.Writer
}

func (obj *writeCloser) Close() error {
	return nil
}
func (obj *writeCloser) Write(p []byte) (n int, err error) {
	return obj.Writer.Write(p)
}
func NoClose(w io.Writer) io.WriteCloser {
	return &writeCloser{Writer: w}
}

// 创建json解析器
func NewJsClient(pre_ctx context.Context, options ...JsClientOption) (*JyClient, error) {
	var option JsClientOption
	if len(options) > 0 {
		option = options[0]
	}
	if option.NodePath == "" {
		option.NodePath = "node"
	}
	nowDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if option.ModulePath == nil {
		option.ModulePath = []string{nowDir}
	} else {
		option.ModulePath = append(option.ModulePath, nowDir)
	}
	userDir, err := conf.GetMainDirPath()
	if err != nil {
		return nil, err
	}
	filePath := tools.PathJoin(userDir, fmt.Sprintf(".cmdPipJsScript%s.js", scriptVersion))
	if !tools.PathExist(filePath) {
		err := os.WriteFile(filePath, cmdPipJsScript, 0777)
		if err != nil {
			return nil, err
		}
	}
	cli, err := NewClient(pre_ctx, ClientOption{
		Dir:  nowDir,
		Name: option.NodePath,
		Args: []string{filePath},
	})
	if err != nil {
		return nil, err
	}
	writeBody, err := cli.StdInPipe()
	if err != nil {
		return nil, err
	}
	readPip, err := cli.StdOutPipe()
	if err != nil {
		return nil, err
	}
	go cli.Run()
	jsCli := &JyClient{
		client: cli,
		read:   bufio.NewReader(readPip),
		write:  writeBody,
	}
	return jsCli, jsCli.init(pre_ctx, option.ModulePath)
}

var crlpstart = []byte("##gospider@start##")
var crlpend = []byte("##gospider@end##")

func (obj *JyClient) response() (*gson.Client, error) {
	allCon := []byte{}
	for {
		con, err := obj.read.ReadBytes('#')
		if err != nil {
			return nil, err
		}
		allCon = append(allCon, con...)
		if bytes.HasSuffix(allCon, crlpend) {
			if !bytes.Contains(allCon, crlpstart) {
				return nil, errors.New("数据包不完整")
			}
			startI := bytes.Index(allCon, crlpstart)
			endI := bytes.Index(allCon, crlpend)
			data := allCon[startI+len(crlpstart) : endI]
			return gson.Decode(data)
		}
	}
}
func (obj *JyClient) run(preCtx context.Context, dataMap map[string]any) (jsonData *gson.Client, err error) {
	var ctx context.Context
	var cnl context.CancelFunc
	if preCtx == nil {
		ctx, cnl = context.WithCancel(obj.client.Ctx())
	} else {
		ctx, cnl = context.WithCancel(preCtx)
	}
	defer cnl()
	obj.lock.Lock()
	defer obj.lock.Unlock()
	select {
	case <-ctx.Done():
		if obj.client.Err() != nil {
			return nil, obj.client.Err()
		}
		return nil, errors.New("client closed")
	default:
	}
	con, err := json.Marshal(dataMap)
	if err != nil {
		return nil, err
	}
	con = append(con, '\n')
	if _, err = obj.write.Write(con); err != nil {
		if obj.client.Err() != nil {
			return nil, obj.client.Err()
		}
		return nil, err
	}
	done := make(chan struct{})
	var respErr error
	go func() {
		defer close(done)
		jsonData, respErr = obj.response()
	}()
	select {
	case <-done:
		if obj.client.Err() != nil {
			return nil, obj.client.Err()
		}
		if respErr != nil {
			return nil, respErr
		}
		if jsonData.Get("Error").Exists() && jsonData.Get("Error").String() != "" {
			return jsonData, errors.New(jsonData.Get("Error").String())
		}
		return jsonData.Get("Result"), nil
	case <-ctx.Done():
		if obj.client.Err() != nil {
			return nil, obj.client.Err()
		}
		return nil, context.Cause(ctx)
	}
}

// 执行函数,第一个参数是要调用的函数名称,后面的是传参
func (obj *JyClient) Call(ctx context.Context, funcName string, values ...any) (jsonData *gson.Client, err error) {
	return obj.run(ctx, map[string]any{"Type": "call", "Func": funcName, "Args": values})
}
func (obj *JyClient) Exec(ctx context.Context, script string) (err error) {
	_, err = obj.run(ctx, map[string]any{"Type": "exec", "Script": tools.Base64Encode(script)})
	return
}
func (obj *JyClient) Eval(ctx context.Context, script string) (jsonData *gson.Client, err error) {
	return obj.run(ctx, map[string]any{"Type": "eval", "Script": tools.Base64Encode(script)})
}
func (obj *JyClient) init(ctx context.Context, modulePath ...[]string) (err error) {
	_, err = obj.run(ctx, map[string]any{"Type": "init", "ModulePath": modulePath})
	return
}

// 关闭解析器
func (obj *JyClient) Close() {
	obj.client.Close()
}

// 运行命令
func (obj *Client) Output() ([]byte, error) {
	defer obj.Close()
	return obj.cmd.CombinedOutput()
}

// 运行命令
func (obj *Client) Run() error {
	defer obj.Close()
	err := obj.cmd.Run()
	if err != nil {
		obj.err = err
		return obj.err
	} else if !obj.cmd.ProcessState.Success() {
		if obj.ctx.Err() != nil {
			obj.err = context.Cause(obj.ctx)
			return obj.err
		} else {
			obj.err = errors.New("shell 执行异常")
			return obj.err
		}
	}
	return obj.err
}

// 导出cmd 的 in管道
func (obj *Client) StdInPipe() (io.WriteCloser, error) {
	return obj.cmd.StdinPipe()
}
func (obj *Client) Err() error {
	if obj.cmd.Err != nil {
		return obj.cmd.Err
	}
	return obj.err
}

// 导出cmd 的 out管道
func (obj *Client) StdOutPipe() (io.ReadCloser, error) {
	return obj.cmd.StdoutPipe()
}

// 导出cmd 的error管道
func (obj *Client) StdErrPipe() (io.ReadCloser, error) {
	return obj.cmd.StderrPipe()
}

// 设置cmd 的 error管道
func (obj *Client) SetStdErr(stderr io.WriteCloser) {
	obj.cmd.Stderr = stderr
}

// 设置cmd 的 out管道
func (obj *Client) SetStdOut(stdout io.WriteCloser) {
	obj.cmd.Stdout = stdout
}

// 设置cmd 的 in管道
func (obj *Client) SetStdIn(stdin io.ReadCloser) {
	obj.cmd.Stdin = stdin
}

// 等待运行结束
func (obj *Client) Join() {
	<-obj.ctx.Done()
}

// 关闭客户端
func (obj *Client) Close() {
	defer obj.cnl()
	if obj.cmd.Process != nil {
		killProcess(obj.cmd)
	}
	if obj.closeCallBack != nil {
		obj.closeCallBack()
	}
}

func (obj *Client) Ctx() context.Context {
	return obj.ctx
}
