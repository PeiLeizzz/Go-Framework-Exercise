module github.com/PeiLeizzz/go-gin-example

go 1.18

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/astaxie/beego v1.12.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fvbock/endless v0.0.0-20170109170031-447134032cb6
	github.com/gin-gonic/gin v1.7.7
	github.com/go-ini/ini v1.66.4
	github.com/swaggo/gin-swagger v1.2.0
	github.com/swaggo/swag v1.6.5
	github.com/unknwon/com v1.0.1
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.5 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/tools v0.1.10 // indirect
)

require (
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.10.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/PeiLeizzz/go-gin-example/conf => ../go-gin-example/conf
	github.com/PeiLeizzz/go-gin-example/docs => ../go-gin/example/docs
	github.com/PeiLeizzz/go-gin-example/middleware => ../go-gin-example/middleware
	github.com/PeiLeizzz/go-gin-example/models => ../go-gin-example/models
	github.com/PeiLeizzz/go-gin-example/pkg/e => ../go-gin-example/pkg/e
	github.com/PeiLeizzz/go-gin-example/pkg/setting => ../go-gin-example/pkg/setting
	github.com/PeiLeizzz/go-gin-example/pkg/util => ../go-gin-example/pkg/util
	github.com/PeiLeizzz/go-gin-example/routers => ../go-gin-example/routers
)
