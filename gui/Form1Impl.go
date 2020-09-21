package gui

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"../Conf"
	"../Proxy"
	"../ProxyEntry"
	"github.com/hpcloud/tail"
	"github.com/ying32/govcl/vcl"
	"github.com/ying32/govcl/vcl/types"
)

//::private::
type TForm1Fields struct {
}

var c = make(chan int)

// 启动代理
func (f *TForm1) OnButton1Click(sender vcl.IObject) {
	var UseProxyPool bool = true
	if !Form1.RadioButton1.Checked() {
		UseProxyPool = false
	}
	var PPIP string = Form1.Edit1.Text()
	var PPPort string = Form1.Edit2.Text()
	var Port string = Form1.Edit3.Text()
	// 临时
	// PPIP = "http://10.103.91.179"
	// PPPort = "5010"
	// Port = "8081"
	var UseProxy bool = true
	var UseHttpsProxy bool = true
	var MinProxyNum, _ = strconv.Atoi(Form1.Edit4.Text())
	Conf.InitConfig(MinProxyNum, UseProxyPool, Port, UseProxy, UseHttpsProxy, PPIP, PPPort)
	// 启动代理
	// go ProxyEntry.Proxymain()
	// 可以停止
	go ProxyEntry.Proxymain(c)
	// go func(stop chan int) {
	// 	var isFirst = true
	// 	for {
	// 		select {
	// 		case <-stop:
	// 			stop <- 1
	// 			break
	// 		default:
	// 		}
	// 		if isFirst {
	// 			isFirst = false
	// 			ProxyEntry.Proxymain(stop)
	// 		}
	// 	}
	// }(c)
	// 渲染可用代理池
	go RenderValidProxyPool()

	// 启动日志实时输出
	//go logRealTime()//拉低性能，暂时取消
}

// 停止代理
func (f *TForm1) OnButton2Click(sender vcl.IObject) {
	c <- 1
	<-c
	log.Println("停止代理")
}

func (f *TForm1) OnButton3Click(sender vcl.IObject) {
	dlgOpen := vcl.NewOpenDialog(f)
	dlgOpen.SetFilter("代理列表(*.ls)|*.lst|文本文件(*.txt)|*.txt|所有文件(*.*)|*.*")
	dlgOpen.SetOptions(dlgOpen.Options().Include(types.OfShowHelp, types.OfAllowMultiSelect))
	dlgOpen.SetTitle("打开")
	// 打开文件成功后
	if dlgOpen.Execute() {
		f.ListBox2.Items().Add(time.Now().Format(fmt.Sprintf("2006-01-02 15:04:05 : %s", "导入代理文件")))
		log.Println("导入文件")
		log.Println(dlgOpen.FileName())
		Conf.CustomProxyFile = dlgOpen.FileName()
		var tmpmap = make(map[string]Proxy.Aproxy)
		tmpmap = Proxy.GetMetaproxyFromFile()
		Proxy.MetaProxymap = tmpmap
		f.ListView1.Items().BeginUpdate()
		i := 0
		for k := range tmpmap {
			i++
			item := f.ListView1.Items().Add()
			item.SetCaption(fmt.Sprintf("%d", i))
			item.SubItems().Add(tmpmap[k].Protocol)
			item.SubItems().Add(tmpmap[k].Ip)
			item.SubItems().Add(tmpmap[k].Port)
		}
		f.ListView1.Items().EndUpdate()

	}

}

// 保存
func (f *TForm1) OnButton4Click(sender vcl.IObject) {
	// 判断listview中是否有内容
	if f.ListView1.Items().Count() < 1 {
		return
	}
	// 打开文件
	file, err := os.OpenFile(Conf.SaveProxyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Println("打开文件失败，err: ", err.Error())
		return
	}
	defer file.Close()
	var i int32
	for i = 0; i < f.ListView1.Items().Count(); i++ {
		item := f.ListView1.Items().Item(i)
		protocol := item.SubItems().ValueFromIndex(0)
		ip := item.SubItems().ValueFromIndex(1)
		port := item.SubItems().ValueFromIndex(2)
		line := protocol + "," + ip + ":" + port
		file.WriteString(line + "\n")
		// log.Println(line)
	}
}

// 添加代理
func (f *TForm1) OnButton5Click(sender vcl.IObject) {
	Form2.ShowModal()
}

// 删除代理
func (f *TForm1) OnButton6Click(sender vcl.IObject) {
	protocol := f.ListView1.Selected().SubItems().ValueFromIndex(0)
	ip := f.ListView1.Selected().SubItems().ValueFromIndex(1)
	port := f.ListView1.Selected().SubItems().ValueFromIndex(2)
	delete(Proxy.MetaProxymap, fmt.Sprintf("%x", md5.Sum([]byte(protocol+"://"+ip+":"+port))))
	log.Println(Proxy.MetaProxymap)
	f.ListView1.DeleteSelected()
}

// 日志实时输出
func logRealTime() {
	t, _ := tail.TailFile("log.txt", tail.Config{Follow: true})
	for line := range t.Lines {
		// fmt.Println(line.Text)
		if Form1.ListBox2.Items().Count() > 30 {
			Form1.ListBox2.Items().Delete(0)
		}
		Form1.ListBox2.Items().Add(line.Text)
	}
}

// 定时渲染可用代理池
func RenderValidProxyPool() {
	ticker := time.NewTicker(time.Duration(2 * time.Second))
	for range ticker.C {
		Form1.ListBox1.Items().Clear()
		for k := range Proxy.Proxymap {
			tmp := Proxy.Proxymap[k]
			Form1.ListBox1.Items().Add(tmp.Protocol + "://" + tmp.Ip + ":" + tmp.Port)
		}
	}

}

func (f *TForm1) OnButton7Click(sender vcl.IObject) {
	if f.ListBox1.Items().Count() < 1 {
		return
	}
	// 打开文件
	file, err := os.OpenFile(Conf.SaveProxyFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Println("打开文件失败，err: ", err.Error())
		return
	}
	defer file.Close()
	var i int32
	for i = 0; i < f.ListBox1.Items().Count(); i++ {
		item := f.ListBox1.Items().ValueFromIndex(i)
		protocol := strings.Split(item, ":")[0]
		ipport := strings.Split(item, "/")[2]
		file.WriteString(protocol + "," + ipport + "\n")
	}
	log.Println("追加结束")
}

func (f *TForm1) OnButton8Click(sender vcl.IObject) {
	if f.ListBox1.Items().Count() < 1 {
		return
	}
	// 打开文件
	file, err := os.OpenFile(Conf.SaveProxyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Println("打开文件失败，err: ", err.Error())
		return
	}
	defer func() {
		file.Close()
	}()
	var i int32
	for i = 0; i < f.ListBox1.Items().Count(); i++ {
		item := f.ListBox1.Items().ValueFromIndex(i)
		protocol := strings.Split(item, ":")[0]
		ipport := strings.Split(item, "/")[2]
		file.WriteString(protocol + "," + ipport + "\n")
	}
	log.Println("覆盖结束")
}

func (f *TForm1) OnMenuItem2Click(sender vcl.IObject) {

}

func (f *TForm1) OnMenuItem3Click(sender vcl.IObject) {

}

func (f *TForm1) OnMenuItem4Click(sender vcl.IObject) {

}

func (f *TForm1) OnMenuItem5Click(sender vcl.IObject) {

}