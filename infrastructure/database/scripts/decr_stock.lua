local stockKey = KEYS[1]
local soldoutKey = KEYS[2]
local orderKey = KEYS[3]

-- 返回码说明:
--  1  = 成功（库存扣减 + 未售罄）
-- -1  = 售罄
-- -2  = 库存不存在（未预热）
-- -3  = 库存异常（<=0 但未标记售罄）
-- -4  = 该用户已下单（防重复秒杀）

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

return 1

