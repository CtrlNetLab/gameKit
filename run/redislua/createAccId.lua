if #ARGV ~= 1 then
    return 0
end

if tonumber(ARGV[1])>1023 then
    return 0
end

local AreaId = ARGV[1]
local bit_idServiceId = 15
local bit_idType = 5
local IDType_Player = 1

local seed = tonumber(redis.call("INCR", "DontDelCreateAccId"))

-- accid奇偶性质与seed相关
if seed % 2 == 0 then
    IDType_Player = 2
end

local accId = (seed * 2^bit_idServiceId) + AreaId
accId = (accId * 2^bit_idType) + IDType_Player
local accIdStr = string.format("%.0f", accId)

local key = "DontDelUsedAccId:" .. accIdStr
local exist = redis.call("EXISTS", key)

-- 没有被占用，设置标记
if exist ~= 1 then
    redis.call("SET", key, "1")
    return accId
end

return 0
