package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
)

func TestDelKeyFail(t *testing.T) {
	var (
		ctx    = context.TODO()
		key    = "test"
		fields = "test"
		setErr = errors.New("Failed to delete existing hash. Redis could be empty\n")
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectHDel(key, fields).SetErr(setErr)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	err := client.DelHashKey(key, fields)
	if err.Error() != setErr.Error() {
		t.Error("expectation error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestDelKeySuccess(t *testing.T) {
	var (
		ctx    = context.TODO()
		key    = "test"
		fields = map[string]interface{}{
			"test": "test",
		}
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.Regexp().ExpectHSet(key, fields, `^[a-z]+$`).SetVal(1)
	mock.ExpectHDel(key).SetVal(1)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	_ = client.HashSet(key, fields)
	err := client.DelHashKey(key)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestHashSetSuccess(t *testing.T) {
	var (
		ctx   = context.TODO()
		key   = "test"
		value = map[string]interface{}{
			"test": "test",
		}
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.Regexp().ExpectHSet(key, value, `^[a-z]+$`).SetVal(1)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	err := client.HashSet(key, value)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRightPushSuccess(t *testing.T) {
	var (
		ctx   = context.TODO()
		key   = "test"
		value = "test"
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectRPush(key, value).SetVal(1)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	err := client.RightPush(key, value)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestListLengthSuccess(t *testing.T) {
	var (
		ctx = context.TODO()
		key = "test"
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectLLen(key).SetVal(1)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	_, err := client.ListLength(key)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestLeftPopSuccess(t *testing.T) {
	var (
		ctx = context.TODO()
		key = "test"
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectLPop(key).SetVal(key)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	_, err := client.LeftPop(key)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestHashGetAllSuccess(t *testing.T) {
	var (
		ctx   = context.TODO()
		key   = "test"
		value = map[string]string{
			"test": "test",
		}
	)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.ExpectHGetAll(key).SetVal(value)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	_, err := client.HashGetAll(key)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestExpireKeySuccess(t *testing.T) {
	var (
		ctx    = context.TODO()
		key    = "test"
		fields = map[string]interface{}{
			"test": "test",
		}
	)

	var expireInSeconds time.Duration
	expireInSeconds = 100
	duration := (time.Duration(expireInSeconds) * time.Second)

	redis, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(true)
	mock.Regexp().ExpectHSet(key, fields, `^[a-z]+$`).SetVal(1)
	mock.ExpectExpire(key, duration).SetVal(true)

	client := Client{
		Redis:   redis,
		Context: ctx,
	}

	_ = client.HashSet(key, fields)
	err := client.Expire(key, duration)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
