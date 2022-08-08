package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/kuznetsovin/egts-protocol/cli/receiver/config"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/server"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage"
	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {

	test_conf := os.Getenv("TEST_CONFIG")
	assert.NotEmpty(t, test_conf)

	conf, err := config.New(test_conf)
	if !assert.NoError(t, err) {
		return
	}

	pg_conf, ok := conf.Store["postgresql"]
	if !assert.True(t, ok) {
		return
	}

	pg_store, err := initTestPostgresql(pg_conf)
	if !assert.NoError(t, err) {
		return
	}

	redis_conf, ok := conf.Store["redis"]
	if !assert.True(t, ok) {
		return
	}

	redis_store := initTestRedis(redis_conf)

	mysql_conf, ok := conf.Store["mysql"]
	if !assert.True(t, ok) {
		return
	}

	mysql_store, err := initTestMysql(mysql_conf)
	if !assert.NoError(t, err) {
		return
	}

	storages := storage.NewRepository()
	err = storages.LoadStorages(conf.Store)
	if !assert.NoError(t, err) {
		return
	}

	go func() {
		srv := server.New(conf.GetListenAddress(), conf.GetEmptyConnTTL(), storages)
		srv.Run()
	}()
	time.Sleep(time.Second)

	message := []byte{0x01, 0x00, 0x00, 0x0B, 0x00, 0xB1, 0x00, 0xE8, 0x04, 0x01, 0x4E, 0xA6, 0x00, 0xA1, 0x0A, 0x81, 0x34, 0xF6, 0xE9, 0x01,
		0x02, 0x02, 0x10, 0x1A, 0x00, 0x4F, 0x5F, 0xE5, 0x10, 0x00, 0xBE, 0xCD, 0x9E, 0x80, 0x7F, 0x8B, 0x35, 0x93, 0x9B, 0x80, 0x2F, 0xF9, 0x80,
		0x02, 0x01, 0x00, 0x92, 0x00, 0x00, 0x00, 0x00, 0x11, 0x06, 0x00, 0x0E, 0x46, 0x00, 0x00, 0x00, 0x0C, 0x12, 0x1C, 0x00, 0x01, 0x0F, 0xFF,
		0x01, 0x44, 0x6D, 0x00, 0xB8, 0x00, 0x00, 0x0B, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x14, 0x05, 0x00, 0x02, 0xFF, 0x00, 0x29, 0x04, 0x1B, 0x07, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1B, 0x07, 0x00,
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1B, 0x07, 0x00, 0x03, 0x01, 0x00, 0x5A, 0x08, 0x00, 0x00, 0x1B, 0x07, 0x00, 0x04, 0x02, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x19, 0x04, 0x00, 0x64, 0x77, 0x2A, 0x04, 0x19, 0x04, 0x00, 0x65, 0x00, 0x00, 0x00, 0x19, 0x04, 0x00, 0x66, 0x01,
		0x00, 0x00, 0x19, 0x04, 0x00, 0x67, 0x77, 0x2A, 0x04, 0x19, 0x04, 0x00, 0x68, 0x77, 0x2A, 0x04, 0x19, 0x04, 0x00, 0x69, 0x4F, 0x9A, 0x22,
		0x19, 0x04, 0x00, 0x6E, 0x77, 0x2A, 0x04, 0x41, 0xF6}
	response := []byte{0x1, 0x0, 0x0, 0xb, 0x0, 0x10, 0x0, 0xe9, 0x4, 0x0, 0xa1, 0xe8, 0x4, 0x0, 0x6, 0x0, 0x1, 0x0, 0x20, 0x2, 0x2, 0x0, 0x3,
		0x0, 0xa1, 0xa, 0x0, 0x5e, 0xb6}

	conn, err := net.Dial("tcp", conf.GetListenAddress())
	if assert.NoError(t, err) {
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, _ = conn.Write(message)

		buf := make([]byte, 29)
		_, _ = conn.Read(buf)

		assert.Equal(t, response, buf)

	}
	defer conn.Close()

	if res, err := pg_store.pointCount(); assert.NoError(t, err) {
		assert.Equal(t, 1, res)
	}

	if err = redis_store.receivedPoint(); !assert.NoError(t, err) {
		return
	}

	if res, err := mysql_store.pointCount(); assert.NoError(t, err) {
		assert.Equal(t, 1, res)
	}
}
