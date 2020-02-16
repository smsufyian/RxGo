package rxgo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Amb1(t *testing.T) {
	obs := Amb([]Observable{testObservable(1, 2, 3), Empty()})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3))
}

func Test_Amb2(t *testing.T) {
	obs := Amb([]Observable{Empty(), testObservable(1, 2, 3), Empty(), Empty()})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3))
}

func Test_CombineLatest(t *testing.T) {
	obs := CombineLatest(func(ii ...interface{}) interface{} {
		sum := 0
		for _, v := range ii {
			sum += v.(int)
		}
		return sum
	}, []Observable{testObservable(1, 2), testObservable(10, 11)})
	Assert(context.Background(), t, obs, HasSomeItems())
}

func Test_CombineLatest_Empty(t *testing.T) {
	obs := CombineLatest(func(ii ...interface{}) interface{} {
		sum := 0
		for _, v := range ii {
			sum += v.(int)
		}
		return sum
	}, []Observable{testObservable(1, 2), Empty()})
	Assert(context.Background(), t, obs, HasNoItems())
}

func Test_CombineLatest_Error(t *testing.T) {
	obs := CombineLatest(func(ii ...interface{}) interface{} {
		sum := 0
		for _, v := range ii {
			sum += v.(int)
		}
		return sum
	}, []Observable{testObservable(1, 2), testObservable(errFoo)})
	Assert(context.Background(), t, obs, HasNoItems(), HasRaisedError(errFoo))
}

func Test_Concat_SingleObservable(t *testing.T) {
	obs := Concat([]Observable{testObservable(1, 2, 3)})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3))
}

func Test_Concat_TwoObservables(t *testing.T) {
	obs := Concat([]Observable{testObservable(1, 2, 3), testObservable(4, 5, 6)})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3, 4, 5, 6))
}

func Test_Concat_MoreThanTwoObservables(t *testing.T) {
	obs := Concat([]Observable{testObservable(1, 2, 3), testObservable(4, 5, 6), testObservable(7, 8, 9)})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3, 4, 5, 6, 7, 8, 9))
}

func Test_Concat_EmptyObservables(t *testing.T) {
	obs := Concat([]Observable{Empty(), Empty(), Empty()})
	Assert(context.Background(), t, obs, HasNoItem())
}

func Test_Concat_OneEmptyObservable(t *testing.T) {
	obs := Concat([]Observable{Empty(), testObservable(1, 2, 3)})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3))

	obs = Concat([]Observable{testObservable(1, 2, 3), Empty()})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3))
}

func Test_Empty(t *testing.T) {
	obs := Empty()
	Assert(context.Background(), t, obs, HasNoItems())
}

func Test_FromChannel(t *testing.T) {
	ch := make(chan Item)
	go func() {
		ch <- FromValue(1)
		ch <- FromValue(2)
		ch <- FromValue(3)
		close(ch)
	}()
	obs := FromChannel(ch)
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
}

func Test_FromChannel_SimpleCapacity(t *testing.T) {
	ch := FromChannel(make(chan Item, 10)).Observe(WithBufferedChannel(11))
	assert.Equal(t, 10, cap(ch))
}

func Test_FromChannel_ComposedCapacity(t *testing.T) {
	obs1 := FromChannel(make(chan Item, 10)).
		Map(func(_ interface{}) (interface{}, error) {
			return 1, nil
		}, WithBufferedChannel(11))
	assert.Equal(t, 11, cap(obs1.Observe(WithBufferedChannel(13))))

	obs2 := obs1.Map(func(_ interface{}) (interface{}, error) {
		return 1, nil
	}, WithBufferedChannel(12))
	assert.Equal(t, 12, cap(obs2.Observe(WithBufferedChannel(13))))
}

func Test_FromFuncs(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		next <- FromValue(3)
		done()
	})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
}

func Test_FromFuncs_Multiple(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		done()
	}, func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(10)
		next <- FromValue(20)
		done()
	})
	Assert(context.Background(), t, obs, HasItemsNoParticularOrder(1, 2, 10, 20), HasNotRaisedError())
}

func Test_FromFuncs_Close(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		next <- FromValue(3)
		done()
	})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
}

func Test_FromFuncs_SingleDup(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		next <- FromValue(3)
		done()
	})
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
}

func Test_FromFuncs_ComposedDup(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		next <- FromValue(3)
		done()
	}).Map(func(i interface{}) (_ interface{}, _ error) {
		return i.(int) + 1, nil
	}).Map(func(i interface{}) (_ interface{}, _ error) {
		return i.(int) + 1, nil
	})
	Assert(context.Background(), t, obs, HasItems(3, 4, 5), HasNotRaisedError())
	Assert(context.Background(), t, obs, HasItems(3, 4, 5), HasNotRaisedError())
}

func Test_FromFuncs_ComposedDup_EagerObservation(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		next <- FromValue(3)
		done()
	}).Map(func(i interface{}) (_ interface{}, _ error) {
		return i.(int) + 1, nil
	}, WithEagerObservation()).Map(func(i interface{}) (_ interface{}, _ error) {
		return i.(int) + 1, nil
	})
	Assert(context.Background(), t, obs, HasItems(3, 4, 5), HasNotRaisedError())
	// In the case of an eager observation, we already consumed the items produced by FromFuncs
	// So if we create another subscription, it will be empty
	Assert(context.Background(), t, obs, HasNoItem(), HasNotRaisedError())
}

func Test_FromFuncs_Error(t *testing.T) {
	obs := FromFuncs(func(ctx context.Context, next chan<- Item, done func()) {
		next <- FromValue(1)
		next <- FromValue(2)
		next <- FromError(errFoo)
		done()
	})
	Assert(context.Background(), t, obs, HasItems(1, 2), HasRaisedError(errFoo))
}

func Test_FromFuncs_SimpleCapacity(t *testing.T) {
	ch := FromFuncs(func(_ context.Context, _ chan<- Item, done func()) {
		done()
	}).Observe(WithBufferedChannel(5))
	assert.Equal(t, 5, cap(ch))
}

func Test_FromFuncs_ComposedCapacity(t *testing.T) {
	obs1 := FromFuncs(func(_ context.Context, _ chan<- Item, done func()) {
		done()
	}).Map(func(_ interface{}) (interface{}, error) {
		return 1, nil
	}, WithBufferedChannel(11))
	assert.Equal(t, 11, cap(obs1.Observe(WithBufferedChannel(13))))

	obs2 := obs1.Map(func(_ interface{}) (interface{}, error) {
		return 1, nil
	}, WithBufferedChannel(12))
	assert.Equal(t, 12, cap(obs2.Observe(WithBufferedChannel(13))))
}

func Test_FromItem(t *testing.T) {
	single := JustItem(FromValue(1))
	Assert(context.Background(), t, single, HasItem(1), HasNotRaisedError())
	Assert(context.Background(), t, single, HasItem(1), HasNotRaisedError())
}

func Test_FromItems(t *testing.T) {
	obs := Just(FromValue(1), FromValue(2), FromValue(3))
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
	Assert(context.Background(), t, obs, HasItems(1, 2, 3), HasNotRaisedError())
}

func Test_FromItems_SimpleCapacity(t *testing.T) {
	ch := Just(FromValue(1)).Observe(WithBufferedChannel(5))
	assert.Equal(t, 5, cap(ch))
}

func Test_FromItems_ComposedCapacity(t *testing.T) {
	obs1 := Just(FromValue(1)).Map(func(_ interface{}) (interface{}, error) {
		return 1, nil
	}, WithBufferedChannel(11))
	assert.Equal(t, 11, cap(obs1.Observe(WithBufferedChannel(13))))

	obs2 := obs1.Map(func(_ interface{}) (interface{}, error) {
		return 1, nil
	}, WithBufferedChannel(12))
	assert.Equal(t, 12, cap(obs2.Observe(WithBufferedChannel(13))))
}

func Test_FromEventSource_ObservationAfterAllSent(t *testing.T) {
	const max = 10
	next := make(chan Item, max)
	obs := FromEventSource(next, WithBackPressureStrategy(Drop))

	go func() {
		for i := 0; i < max; i++ {
			next <- FromValue(i)
		}
		close(next)
	}()
	time.Sleep(50 * time.Millisecond)

	Assert(context.Background(), t, obs, CustomPredicate(func(items []interface{}) error {
		if len(items) != 0 {
			return errors.New("items should be nil")
		}
		return nil
	}))
}

func Test_FromEventSource_Drop(t *testing.T) {
	const max = 100_000
	next := make(chan Item, max)
	obs := FromEventSource(next, WithBackPressureStrategy(Drop))

	go func() {
		for i := 0; i < max; i++ {
			next <- FromValue(i)
		}
		close(next)
	}()

	Assert(context.Background(), t, obs, CustomPredicate(func(items []interface{}) error {
		if len(items) == max {
			return errors.New("some items should be dropped")
		}
		if len(items) == 0 {
			return errors.New("no items")
		}
		return nil
	}))
}

func Test_Observable_Interval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	obs := Interval(WithDuration(time.Nanosecond), WithContext(ctx))
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	Assert(context.Background(), t, obs, HasSomeItems())
}

func Test_Observable_Interval_NoItem(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	obs := Interval(WithDuration(time.Nanosecond), WithContext(ctx))
	time.Sleep(50 * time.Millisecond)
	cancel()
	// As Interval is built on an event source, we expect no items
	Assert(context.Background(), t, obs, HasNoItem())
}

func Test_Merge(t *testing.T) {
	obs := Merge([]Observable{testObservable(1, 2), testObservable(3, 4)})
	Assert(context.Background(), t, obs, HasItemsNoParticularOrder(1, 2, 3, 4))
}

func Test_Merge_Error(t *testing.T) {
	obs := Merge([]Observable{testObservable(1, 2), testObservable(3, errFoo)})
	// The content is not deterministic, hence we just test if we have some items
	Assert(context.Background(), t, obs, HasSomeItems(), HasRaisedError(errFoo))
}

func Test_Range(t *testing.T) {
	obs := Range(5, 3)
	Assert(context.Background(), t, obs, HasItems(5, 6, 7, 8))
	// Test whether the observable is reproducible
	Assert(context.Background(), t, obs, HasItems(5, 6, 7, 8))
}

func Test_Range_NegativeCount(t *testing.T) {
	obs := Range(1, -5)
	Assert(context.Background(), t, obs, HasRaisedAnError())
}

func Test_Range_MaximumExceeded(t *testing.T) {
	obs := Range(1<<31, 1)
	Assert(context.Background(), t, obs, HasRaisedAnError())
}

func Test_Start(t *testing.T) {
	obs := Start([]Supplier{func(ctx context.Context) Item {
		return FromValue(1)
	}, func(ctx context.Context) Item {
		return FromValue(2)
	}})
	Assert(context.Background(), t, obs, HasItemsNoParticularOrder(1, 2))
}

func Test_Timer(t *testing.T) {
	obs := Timer(WithDuration(time.Nanosecond))
	Assert(context.Background(), t, obs, HasSomeItems())
}

func Test_Timer_Empty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	obs := Timer(WithDuration(time.Hour), WithContext(ctx))
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	Assert(context.Background(), t, obs, HasNoItems())
}