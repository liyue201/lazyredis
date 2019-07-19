package redis
 
import (
	"encoding/hex"
	"errors"
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/wuxibin89/redis-go-cluster"
)

type Client interface {
	Do(cmd string, args ...interface{}) (string, error)
	Close() error
}

func NewRedisClient(addr [] string, password string, db int) (Client, error) {
	if len(addr) == 0 {
		return nil, errors.New("addr is empty")
	}
	if len(addr) > 1 {
		return NewClusterClient(&redis.Options{StartNodes: addr})
	}
	return NewSimpleClient(addr[0], redigo.DialPassword(password), redigo.DialDatabase(db))
}

func replyText(reply interface{}) string {
	if n, err := redigo.Int(reply, nil); err == nil {
		return fmt.Sprintf("%d", n)
	}
	if n, err := redigo.Int64(reply, nil); err == nil {
		return fmt.Sprintf("%d", n)
	}
	if n, err := redigo.Uint64(reply, nil); err == nil {
		return fmt.Sprintf("%d", n)
	}
	if n, err := redigo.Float64(reply, nil); err == nil {
		return fmt.Sprintf("%f", n)
	}
	if str, err := redigo.String(reply, nil); err == nil {
		return str
	}
	if bys, err := redigo.Bytes(reply, nil); err == nil {
		return hex.EncodeToString(bys)
	}
	if b, err := redigo.Bool(reply, nil); err == nil {
		if b {
			return "true"
		}
		return "false"
	}
	if values, err := redigo.Values(reply, nil); err == nil {
		str := ""
		for _, v := range values {
			if str != "" {
				str += "\n"
			}
			str += replyText(v)
		}
		return str
	}
	if floats, err := redigo.Float64s(reply, nil); err == nil {
		str := ""
		for _, v := range floats {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s", v)
		}
		return str
	}

	if strs, err := redigo.Strings(reply, nil); err == nil {
		str := ""
		for _, v := range strs {
			if str != "" {
				str += "\n"
			}
			str += v
		}
		return str
	}

	if byteSlids, err := redigo.ByteSlices(reply, nil); err == nil {
		str := ""
		for _, v := range byteSlids {
			if str != "" {
				str += "\n"
			}
			str += hex.EncodeToString(v)
		}
		return str
	}

	if ints, err := redigo.Int64s(reply, nil); err == nil {
		str := ""
		for _, v := range ints {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%d", v)
		}
		return str
	}
	if ints, err := redigo.Ints(reply, nil); err == nil {
		str := ""
		for _, v := range ints {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%d", v)
		}
		return str
	}

	if mps, err := redigo.StringMap(reply, nil); err == nil {
		str := ""
		for k, v := range mps {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s %s", k, v)
		}
		return str
	}

	if mps, err := redigo.IntMap(reply, nil); err == nil {
		str := ""
		for k, v := range mps {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s %d", k, v)
		}
		return str
	}

	if mps, err := redigo.Int64Map(reply, nil); err == nil {
		str := ""
		for k, v := range mps {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s %d", k, v)
		}
		return str
	}

	if pos, err := redigo.Positions(reply, nil); err == nil {
		str := ""
		for _, v := range pos {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%f %f", v[0], v[1])
		}
		return str
	}
	return ""
}
