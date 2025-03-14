cd ../
rm -rf go.mod go.sum
go mod init LOGProcessor
go mod tidy
export SELF_CFG_PATH=${PWD}
cd -