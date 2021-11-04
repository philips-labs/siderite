package iron

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData = `{
 "cluster_info": [
  {
   "cluster_id": "randomcluster",
   "cluster_name": "clustername",
   "pubkey": "-----BEGIN PUBLIC KEY----- MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEArw0rWJiR4ZJwuVoxm4IK GQQe1jjG1Kdn19rMq4K0MUgyP7SdA6vmlolLqtB7sc69vUwH695MIuhTD/EKOuzs d2Hu1FTwUZQR8jDqo6r7+xlEEtspj2KSuxqN+Z/FdrQ34aQmmKMs2YSq3hkrCvdv CZDuZ9v7eCwWqFgv3sBcfNbTkS8nZ6Px0itf4t+sUIFEAXCQtEnpdBBrOnUvrxVS vUVK9s63hOJ18YVT3UDaSgfl2Z5biD4ctgUm/w+oPES3LFQZPcCu2s7w8+mtiWDB 3qngxq+/M6CgiM9iNqkQbwtgutM5iwnok49McqzadM7FfKPjd6MMTGMBKbvFIn4+ yq1TnIGbJDUdQj4plhTdqX+xmdwsvpjn+2NcH9YasvIH6phmpRE1VdYILdWzlZMl e7RDc8+3pDoPHW9XzyQgYORst3MQqfqp5fA0KfJGl0myPB/9QXq0dUGeowNQ1M3v G5x0AzPkPehsXs0uHnKhJvjyObD+SxayNMJAd90wgJ1QV7aYTtg6oIL0lM3e7jjr oqhXvXFnx3WEPx96sv+1TdHHaNdYbfEGl6yeclJfWUjzYB/nFTDfYTYq5Ntg/SPJ 1E7nwwsdPju2Tu1Q135us+aOtoc59LYrlyWiLkTHViDHLnZjPqW/HUPudR64mG1Q 1NPjC1OneqQc6ESEn8Jro0UCAwEAAQ== -----END PUBLIC KEY-----",
   "user_id": "userid"
  }
 ],
 "email": "email@foo.com",
 "password": "strongpassword",
 "project": "someproject",
 "project_id": "projectid1234",
 "token": "t0kenh3re",
 "user_id": "userone"
}`

func TestLoadConfig(t *testing.T) {
	file, err := os.CreateTemp("", "iron.*.json")
	if !assert.Nil(t, err) {
		return
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()
	err = ioutil.WriteFile(file.Name(), []byte(testData), 0600)
	if !assert.Nil(t, err) {
		return
	}
	cfg, err := Load(file.Name())
	if !assert.Nil(t, err) {
		return
	}
	if !assert.NotNil(t, cfg) {
		return
	}
	assert.Equal(t, "t0kenh3re", cfg.Token)
	assert.Equal(t, "userone", cfg.UserID)
}
