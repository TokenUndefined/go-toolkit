package container_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mylxsw/go-toolkit/container"
)

type GetUserInterface interface {
	GetUser() string
}

type GetRoleInterface interface {
	GetRole() string
}

type RoleService struct{}

func (r RoleService) GetRole() string {
	return "admin"
}

type UserService struct {
	repo *UserRepo
}

func (u *UserService) GetUser() string {
	return fmt.Sprintf("get user from connection: %s", u.repo.connStr)
}

type UserRepo struct {
	connStr string
}

var expectedValue = "get user from connection: root:root@/my_db?charset=utf8"

func TestPrototype(t *testing.T) {
	c := container.New()

	c.MustBindValue("conn_str", "root:root@/my_db?charset=utf8")
	c.MustSingleton(func(c *container.Container) (*UserRepo, error) {
		connStr, err := c.Get("conn_str")
		if err != nil {
			return nil, err
		}

		return &UserRepo{connStr: connStr.(string)}, nil
	})
	c.MustPrototype(func(userRepo *UserRepo) *UserService {
		return &UserService{repo: userRepo}
	})

	if err := c.Resolve(func(userService *UserService) {
		if userService.GetUser() != expectedValue {
			t.Error("test failed")
		}
	}); err != nil {
		t.Errorf("test failed: %s", err)
		return
	}
	// reflect.TypeOf((*UserService)(nil))
	{
		userService, err := c.Get(reflect.TypeOf((*UserService)(nil)))
		if err != nil {
			t.Error(err)
			return
		}

		if userService.(*UserService).GetUser() != expectedValue {
			t.Error("test failed")
		}
	}

	{
		userService, err := c.Get((*UserService)(nil))
		if err != nil {
			t.Error(err)
			return
		}

		if userService.(*UserService).GetUser() != expectedValue {
			t.Error("test failed")
		}
	}

	{
		c.MustResolve(func(cc *container.Container) {
			userService, err := c.Get((*UserService)(nil))
			if err != nil {
				t.Error(err)
				return
			}

			if userService.(*UserService).GetUser() != expectedValue {
				t.Error("test failed")
			}
		})
	}
}

func TestInterfaceInjection(t *testing.T) {
	c := container.New()
	c.MustBindValue("conn_str", "root:root@/my_db?charset=utf8")
	c.MustSingleton(func(c *container.Container) (*UserRepo, error) {
		connStr, err := c.Get("conn_str")
		if err != nil {
			return nil, err
		}

		return &UserRepo{connStr: connStr.(string)}, nil
	})
	c.MustPrototype(func(userRepo *UserRepo) (*UserService, error) {
		return &UserService{repo: userRepo}, nil
	})

	if err := c.Resolve(func(userService GetUserInterface) {
		if userService.GetUser() != expectedValue {
			t.Error("test failed")
		}
	}); err != nil {
		t.Errorf("test failed: %s", err)
	}

	c.MustPrototype(func() (RoleService, error) {
		return RoleService{}, nil
	})

	err := c.Resolve(func(roleService GetRoleInterface) {
		if roleService.GetRole() != "admin" {
			t.Error("test failed")
		}
	})
	if err != nil {
		t.Error(err)
	}
}

func TestWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	c := container.NewWithContext(ctx)
	c.MustResolve(func(ctx context.Context) {
		startTime := time.Now().UnixNano()
		<-ctx.Done()

		if (time.Now().UnixNano()-startTime)/1000/1000 < 100 {
			t.Error("test failed")
		}
	})

	time.Sleep(200 * time.Millisecond)
}

type TestObject struct {
	Name string
}

func TestWithProvider(t *testing.T) {
	c := container.New()
	c.MustBindValue("conn_str", "root:root@/my_db?charset=utf8")
	c.MustSingleton(func(c *container.Container) (*UserRepo, error) {
		connStr, err := c.Get("conn_str")
		if err != nil {
			return nil, err
		}

		return &UserRepo{connStr: connStr.(string)}, nil
	})
	c.MustPrototype(func(userRepo *UserRepo) *UserService {
		return &UserService{repo: userRepo}
	})

	provider, err := c.ServiceProvider(func() *TestObject {
		return &TestObject{Name: "mylxsw"}
	})
	if err != nil {
		t.Error("test failed")
	}
	if _, err := c.CallWithProvider(func(userService *UserService, testObject *TestObject) {
		if userService.GetUser() != expectedValue {
			t.Error("test failed")
		}

		if testObject.Name != "mylxsw" {
			t.Error("test failed")
		}
	}, provider); err != nil {
		t.Errorf("test failed: %s", err)
		return
	}
}
