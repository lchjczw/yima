# Yima

> [易码平台](http://51ym.me/)的一个简单包装

搜索短信模版，并且获取一个手机号

```go
package main

import (
	"github.com/wolther47/yima"
	"fmt"
)

func main(){

    y := yima.Yima{Token: "<YOUR TOKEN>"}

    candidates, err := y.SearchTemplate("<KEYWORD>")
    if err != nil {
        panic(err)
    }
    
    // 使用第一个
    itemID := candidates[0].ID
    
    // 获取一个中国移动的号码
    phone, err := y.GetNumber(itemID, &yima.MobileOption{
        ISP: yima.ChinaMobile,
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Get the phone: %v", phone)
    
}
```

## License

Under MIT License