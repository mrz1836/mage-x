module github.com/mrz1836/mage-x/examples/override-commands

go 1.24.9

require github.com/mrz1836/mage-x v1.13.3

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	go.uber.org/mock v0.6.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use local mage-x for development
replace github.com/mrz1836/mage-x => ../../
