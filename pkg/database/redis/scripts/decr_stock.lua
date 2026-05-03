local key = KEYS[1]
local stock = tonumber(redis.call('GET,key'))

if not stock or stock <= 0 then
    return 0
end

redis.call('DECR',key)
return 1