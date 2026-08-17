module github.com/mrz1836/mage-x/examples/override-commands

go 1.25.0

require github.com/mrz1836/mage-x v1.20.6

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mrz1836/go-selfupdate v0.1.3 // indirect
	go.uber.org/mock v0.6.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use local mage-x for development
replace github.com/mrz1836/mage-x => ../../
