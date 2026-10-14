--[[
-- usage: FindPlayer.lua UserId AccId Name
-- retval: {accId}
-- 先精确查找accid， 再模糊查询userid，最后再模糊查询name
]]--

local UserId = ARGV[1]
local AccId = ARGV[2]
local Name = ARGV[3]
local PlatType = ARGV[4]
local Account = ARGV[5]

-- note: 键名前缀已在其他文件中定义
local UserIdKey = "UserIdIndex:" .. UserId
local NameKey = "NameIndex:" .. Name
local AccIdKey = "AccIdIndex:" .. AccId
local AccountKey = "AccountIndex:" .. PlatType .. "_" .. Account

local AccIds = {}
if tonumber(AccId) ~= 0 then
    local accId = redis.call("EXISTS", "DontDelUsedAccId:"..AccId)
    if accId ~= 0 then
        table.insert(AccIds, AccId)
        return AccIds
    end
end

if UserId ~= "" then
    local accId = redis.call("HGET", UserIdKey, "AccId")
    if accId ~= false then
        table.insert(AccIds, accId)
        return AccIds
    end
end

if Name ~= "" then
    local accIds = redis.call('SMEMBERS', NameKey)
    for _, accId in pairs(accIds) do
        table.insert(AccIds, accId)
    end
    return AccIds
end

if PlatType ~= "" and Account ~= "" then
    local accId = redis.call('GET', AccountKey)
    if accId ~= false then
        table.insert(AccIds, accId)
        return AccIds
    end
end

return AccIds
