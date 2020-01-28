module github.com/pburakov/playback

require (
	cloud.google.com/go v0.36.0
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/linkedin/goavro v2.1.0+incompatible
	github.com/stretchr/testify v1.3.0
	github.com/testcontainers/testcontainers-go v0.0.0-20190207081624-4ed65004fe50
	gopkg.in/linkedin/goavro.v1 v1.0.5 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

go 1.12
