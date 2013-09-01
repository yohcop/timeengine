
mkdir -p genfiles/src/timeengine/mock_ae/

third_party/bin/mockgen --destination=genfiles/src/timeengine/mock_ae/mock.go timeengine/ae Context
