package trans

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Seat locks live in two Redis structures per show:
//
//	locks:{session}       ZSET  seat -> lock expiry (unix ms)
//	lockowners:{session}  HASH  seat -> transaction ID
//
// All operations run as Lua scripts so that locking N seats is atomic: either
// every requested seat is locked for the transaction or none is. Expired
// entries are swept lazily on each lock attempt, so an abandoned checkout
// frees its seats without any background job.

var lockScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local expired = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', now)
if #expired > 0 then
  redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now)
  redis.call('HDEL', KEYS[2], unpack(expired))
end
for i = 4, #ARGV do
  if redis.call('ZSCORE', KEYS[1], ARGV[i]) then
    if redis.call('HGET', KEYS[2], ARGV[i]) ~= ARGV[3] then
      return ARGV[i]
    end
  end
end
for i = 4, #ARGV do
  redis.call('ZADD', KEYS[1], ARGV[2], ARGV[i])
  redis.call('HSET', KEYS[2], ARGV[i], ARGV[3])
end
redis.call('PEXPIREAT', KEYS[1], tonumber(ARGV[2]) + 86400000)
redis.call('PEXPIREAT', KEYS[2], tonumber(ARGV[2]) + 86400000)
return ''
`)

var releaseScript = redis.NewScript(`
for i = 2, #ARGV do
  if redis.call('HGET', KEYS[2], ARGV[i]) == ARGV[1] then
    redis.call('ZREM', KEYS[1], ARGV[i])
    redis.call('HDEL', KEYS[2], ARGV[i])
  end
end
return 1
`)

var verifyScript = redis.NewScript(`
local now = tonumber(ARGV[1])
for i = 3, #ARGV do
  local score = redis.call('ZSCORE', KEYS[1], ARGV[i])
  if not score or tonumber(score) <= now then return 0 end
  if redis.call('HGET', KEYS[2], ARGV[i]) ~= ARGV[2] then return 0 end
end
return 1
`)

func locksKey(session int64) string  { return "locks:" + strconv.FormatInt(session, 10) }
func ownersKey(session int64) string { return "lockowners:" + strconv.FormatInt(session, 10) }

// lockSeats returns the first seat that is held by someone else, or "".
func lockSeats(ctx context.Context, rdb *redis.Client, session int64, txn string, seats []string, until time.Time) (string, error) {
	args := []any{time.Now().UnixMilli(), until.UnixMilli(), txn}
	for _, s := range seats {
		args = append(args, s)
	}
	return lockScript.Run(ctx, rdb, []string{locksKey(session), ownersKey(session)}, args...).Text()
}

func releaseSeats(ctx context.Context, rdb *redis.Client, session int64, txn string, seats []string) error {
	args := []any{txn}
	for _, s := range seats {
		args = append(args, s)
	}
	return releaseScript.Run(ctx, rdb, []string{locksKey(session), ownersKey(session)}, args...).Err()
}

func verifySeats(ctx context.Context, rdb *redis.Client, session int64, txn string, seats []string) (bool, error) {
	args := []any{time.Now().UnixMilli(), txn}
	for _, s := range seats {
		args = append(args, s)
	}
	n, err := verifyScript.Run(ctx, rdb, []string{locksKey(session), ownersKey(session)}, args...).Int()
	return n == 1, err
}

// lockedSeats lists seats with a live lock on a show.
func lockedSeats(ctx context.Context, rdb *redis.Client, session int64) ([]string, error) {
	return rdb.ZRangeByScore(ctx, locksKey(session), &redis.ZRangeBy{Min: fmt.Sprintf("(%d", time.Now().UnixMilli()), Max: "+inf"}).Result()
}

// General-admission events hold quantities rather than seats:
//
//	holds:{event}:{tier}    ZSET  txn -> hold expiry (unix ms)
//	holdqty:{event}:{tier}  HASH  txn -> quantity
//
// ARGV[5] carries the committed sales count read from Postgres; the script
// admits the hold only if sold + live holds + requested <= capacity.

var holdScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local expired = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', now)
if #expired > 0 then
  redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now)
  redis.call('HDEL', KEYS[2], unpack(expired))
end
local held = 0
for _, t in ipairs(redis.call('ZRANGE', KEYS[1], 0, -1)) do
  if t ~= ARGV[3] then held = held + tonumber(redis.call('HGET', KEYS[2], t) or '0') end
end
local remaining = tonumber(ARGV[6]) - tonumber(ARGV[5]) - held
if tonumber(ARGV[4]) > remaining then return {0, remaining} end
redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
redis.call('HSET', KEYS[2], ARGV[3], ARGV[4])
return {1, remaining - tonumber(ARGV[4])}
`)

var heldScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local held = 0
for _, t in ipairs(redis.call('ZRANGEBYSCORE', KEYS[1], '(' .. now, '+inf')) do
  held = held + tonumber(redis.call('HGET', KEYS[2], t) or '0')
end
return held
`)

func holdKeys(event, tier string) []string {
	return []string{"holds:" + event + ":" + tier, "holdqty:" + event + ":" + tier}
}

// holdTickets reports whether the hold was placed, and the capacity left.
func holdTickets(ctx context.Context, rdb *redis.Client, event, tier, txn string, qty, sold, capacity int, until time.Time) (bool, int, error) {
	res, err := holdScript.Run(ctx, rdb, holdKeys(event, tier), time.Now().UnixMilli(), until.UnixMilli(), txn, qty, sold, capacity).Int64Slice()
	if err != nil || len(res) != 2 {
		return false, 0, err
	}
	return res[0] == 1, int(res[1]), nil
}

func heldTickets(ctx context.Context, rdb *redis.Client, event, tier string) (int, error) {
	return heldScript.Run(ctx, rdb, holdKeys(event, tier), time.Now().UnixMilli()).Int()
}

func verifyHold(ctx context.Context, rdb *redis.Client, event, tier, txn string) (bool, error) {
	score, err := rdb.ZScore(ctx, holdKeys(event, tier)[0], txn).Result()
	if err == redis.Nil {
		return false, nil
	}
	return err == nil && int64(score) > time.Now().UnixMilli(), err
}

func releaseHold(ctx context.Context, rdb *redis.Client, event, tier, txn string) error {
	k := holdKeys(event, tier)
	pipe := rdb.TxPipeline()
	pipe.ZRem(ctx, k[0], txn)
	pipe.HDel(ctx, k[1], txn)
	_, err := pipe.Exec(ctx)
	return err
}
