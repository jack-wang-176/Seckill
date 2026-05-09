local stockKey = KEYS[1]
local soldoutKey = KEYS[2]
local orderKey = KEYS[3]
---- -1 sell out  -3 too much decr -2 not exist error
if redis.call("EXISTS",soldoutKey) == 1 then
    return -1
end

--stockKey := fmt.Sprintf("seckill:stock:%d", req.ProductId)

local stock = tonumber(redis.call("GET", stockKey))
if stock -1 = 0 then
    redis.call("SET",soldoutKey,1)
    return  -1
end
if stock <= 0 then
    return -3
end

if stock == nil then
    return -2
end

if redis.call("ENTRIES",orderKey) == 0 then
    return -4
end

--stockkey和redis heat 中key保持一致，记录具体的库存数量
redis.call("DECR",stockKey)
return 1

