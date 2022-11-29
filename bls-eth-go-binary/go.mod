module github.com/herumi/bls-eth-go-binary

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.6.1 // indirect
	github.com/dgraph-io/ristretto v0.1.1
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3 // indirect
	github.com/phoreproject/bls v0.0.0-20200525203911-a88a5ae26844 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prysmaticlabs/prysm/v3 v3.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace (
	github.com/360EntSecGroup-Skylar/excelize/v2 => github.com/xuri/excelize/v2 v2.6.0
	github.com/xuri/excelize/v2 => github.com/360EntSecGroup-Skylar/excelize/v2 v2.6.0
)

go 1.12
