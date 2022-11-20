module github.com/herumi/bls-eth-go-binary

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.6.1 // indirect
	github.com/dgraph-io/ristretto v0.1.1
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/360EntSecGroup-Skylar/excelize/v2 => github.com/xuri/excelize/v2 v2.6.0
	github.com/xuri/excelize/v2 => github.com/360EntSecGroup-Skylar/excelize/v2 v2.6.0
)

go 1.12
