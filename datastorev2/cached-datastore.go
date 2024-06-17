package datastorev2

// import (
// 	"cmp"
// 	"context"
// 	"errors"
// 	"time"

// 	"ggithub.com/patrickmn/go-cache"
// 	"github.com/appliedres/cloudy/models"
// 	"github.com/patrickmn/go-cache"
// )

// var _ JsonDataStore = (*CachedDatastore)(nil)

// type CachedDatastore struct {
// 	backer JsonDataStore
// 	cache  *cache.Cache
// }

// func NewCachedDatastore(backer JsonDataStore) *CachedDatastore {
// 	return &CachedDatastore{
// 		backer: backer,
// 		cache:  cache.New(time.Minute*15, time.Minute*60),
// 	}
// }

// // Close implements UntypedJsonDataStore.
// func (c *CachedDatastore) Close(ctx context.Context) error {
// 	c.cache.Flush()
// 	if c.backer != nil {
// 		return c.backer.Close(ctx)
// 	}
// 	return nil
// }

// // Delete implements UntypedJsonDataStore.
// func (c *CachedDatastore) Delete(ctx context.Context, key string) error {
// 	c.cache.Delete(key)
// 	if c.backer != nil {
// 		return c.backer.Delete(ctx, key)
// 	}
// 	return nil
// }

// // Exists implements UntypedJsonDataStore.
// func (c *CachedDatastore) Exists(ctx context.Context, key string) (bool, error) {
// 	_, found := c.cache.Get(key)
// 	if found {
// 		return true, nil
// 	}
// 	if c.backer != nil {
// 		return c.backer.Exists(ctx, key)
// 	}
// 	return false, nil
// }

// // Get implements UntypedJsonDataStore.
// func (c *CachedDatastore) Get(ctx context.Context, key string) (interface{}, error) {
// 	item, found := c.cache.Get(key)
// 	if found {
// 		return item, nil
// 	}
// 	if c.backer != nil {
// 		item, err := c.backer.Get(ctx, key)
// 		if err != nil {
// 			return nil, err
// 		}
// 		c.cache.SetDefault(key, item)
// 		return item, nil
// 	}
// 	return nil, nil
// }

// // GetAll implements UntypedJsonDataStore.
// func (c *CachedDatastore) GetAll(ctx context.Context, page *models.Page) ([]interface{}, *models.Page, error) {
// 	if c.backer != nil {
// 		all, nextpage, err :=  c.backer.GetAll(ctx, page)
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		// for _, item := range all {
// 		// 	c.cache.SetDefault()
// 		// }
// 	}

// 	if c.cache.ItemCount() == 0 {
// 		return nil, nil, nil
// 	}

// 	size := page.PageSize
// 	if size <= 0 {
// 		size = c.cache.ItemCount()
// 	}

// 	rtn := make([]interface{}, size)
// 	skip := page.Skip
// 	foundPageToken := false

// 	cnt := 0
// 	index := -1
// 	for k, v := range c.cache.Items() {
// 		index++
// 		if skip >= index {
// 			continue
// 		}
// 		if page.NextPageToken != nil && !foundPageToken {
// 			foundPageToken = (page.NextPageToken == k)
// 			continue
// 		}

// 		rtn[cnt] = v
// 		cnt++
// 		if cnt > size {
// 			break
// 		}
// 	}
// 	return rtn,
// }

// // OnCreate implements UntypedJsonDataStore.
// func (c *CachedDatastore) SetOnCreate(fn OnCreateFn) {
// 	c.backer.SetOnCreate(fn)
// }

// // Open implements UntypedJsonDataStore.
// func (c *CachedDatastore) Open(ctx context.Context, config interface{}) error {
// 	c.cache.Flush()
// 	if c.backer != nil {
// 		return c.backer.Open(ctx, config)
// 	}
// 	return nil
// }

// // Query implements UntypedJsonDataStore.
// func (c *CachedDatastore) Query(ctx context.Context, query *SimpleQuery) ([]interface{}, error) {
// 	if c.backer != nil {
// 		return c.backer.Query(ctx, query)
// 	}
// 	return nil, errors.New("Query not supported")
// }

// // Save implements UntypedJsonDataStore.
// func (c *CachedDatastore) Save(ctx context.Context, item interface{}, key string) error {
// 	c.cache.SetDefault(key, item)
// 	if c.backer != nil {
// 		return c.backer.Save(ctx, item, key)
// 	}
// 	return nil
// }

// func Max[T cmp.Ordered](vals ...int) T {
// 	if len(vals) == 0 {
// 		return 0
// 	}
// 	current := vals[0]
// 	for i := 1; i < len(vals); i++ {
// 		if vals[i] > current {
// 			current = vals[i]
// 		}
// 	}
// 	return current
// }
