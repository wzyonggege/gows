package storage

/*
redis 的string 类型做key-value cache
*/

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)


type Config struct {
	Addr        string
	Password    string
	Bb          int
	Idle        int
	Active      int
	IdleTimeout time.Duration
}

type RdsStorage struct {
	pool *redis.Pool
}

func NewRdsStorage(c *Config) (Storage, error)  {
	pool := &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: c.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			d, err := redis.Dial("tcp", c.Addr)
			if err != nil {
				return nil, err
			}
			if c.Password != "" {
				if _, err := d.Do("AUTH", c.Password); err != nil {
					_ = d.Close()
					return nil, err
				}
			}
			if c.Bb != 0 {
				if _, err := d.Do("SELECT", c.Bb); err != nil {
					_ = d.Close()
					return nil, err
				}
			}
			return d, err
		},
	}

	r := &RdsStorage{
		pool: pool,
	}

	return r, nil
}

func (s *RdsStorage) Get(key string) (string, error) {
	conn := s.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("Get", key))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return val, nil
}

func (s *RdsStorage) Set(key string, value string) error {
	conn := s.pool.Get()
	defer conn.Close()
	_, err := conn.Do("Set", key, value)
	// TODO set error return old value
	return err
}

func (s *RdsStorage) Delete(key string) error {
	conn := s.pool.Get()
	defer conn.Close()
	_, err := conn.Do("Del", key)
	return err
}

func (s *RdsStorage) Close() error {
	s.pool.Close()
	return nil
}