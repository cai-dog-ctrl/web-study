package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	version   int    `mapstructure:"version"`
	port      string `mapstructure:"port"`
	MysqlConf `mapstructure:"mysql"`
}
type MysqlConf struct {
	host   string `mapstructure:"host"`
	port   int    `mapstructure:"port"`
	dbName string `mapstructure:"dbname"`
}

func main() {
	//建立默认值
	viper.SetDefault("ContentDir", "content")
	//读取配置文件
	viper.SetConfigFile("./config.yaml")  // 指定配置文件路径
	viper.SetConfigName("config")         // 配置文件名称(无扩展名)
	viper.SetConfigType("yaml")           // 如果配置文件的名称中没有扩展名，则需要配置此项
	viper.AddConfigPath("/etc/appname/")  // 查找配置文件所在的路径
	viper.AddConfigPath("$HOME/.appname") // 多次调用以添加多个搜索路径
	viper.AddConfigPath(".")              // 还可以在工作目录中查找配置
	err := viper.ReadInConfig()           // 查找并读取配置文件
	if err != nil {                       // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	//写入配置文件
	//viper.WriteConfig() // 将当前配置写入“viper.AddConfigPath()”和“viper.SetConfigName”设置的预定义路径
	//viper.SafeWriteConfig()
	//viper.WriteConfigAs("/path/to/my/.config")
	//viper.SafeWriteConfigAs("/path/to/my/.config") // 因为该配置文件写入过，所以会报错
	//viper.SafeWriteConfigAs("/path/to/my/.other_config")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数
		fmt.Println("Config file changed:", e.Name)
	})
	//实时检测配置文件的变化
	//r := gin.Default()
	//r.GET("viper", func(c *gin.Context) {
	//	c.String(http.StatusOK, viper.GetString("version"))
	//})
	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		fmt.Println(err)
	}
}
