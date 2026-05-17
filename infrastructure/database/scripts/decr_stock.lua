local stockKey = KEYS[1]
local soldoutKey = KEYS[2]
local orderKey = KEYS[3]

-- 返回码说明:
--  1  = 成功
-- -1  = 售罄
-- -2  = 库存不存在
-- -3  = 库存异常
-- -4  = 该用户已下单

if redis.call("EXISTS", soldoutKey) == 1 then
    return -1
end

local stock = tonumber(redis.call("GET", stockKey))
if stock == nil then
    return -2
end

if stock <= 0 then
    redis.call("SET", soldoutKey, 1)
    return -1
end

if redis.call("EXISTS", orderKey) == 1 then
    return -4
end

redis.call("DECR", stockKey)
if redis.call("GET", stockKey) == "0" then
    redis.call("SET", soldoutKey, 1)
end


redis.call("SET", orderKey, ARGV[1], "EX", 30)

return 1
