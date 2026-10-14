--[[
--  UpdataPlayerIndex.lua UserId AccId Name
--  更新玩家简要数据，用于平台查找玩家
--]]

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

-- <UserId> -> AccId
-- <Name> -> AccId
-- AccId: <UserId, Name>

-- 先取出名字， 
local oriName = redis.call('HGET', AccIdKey, 'Name')
if oriName ~= false then
    redis.call('SREM', "NameIndex:"..oriName, AccId)
end

-- 设置Name->AccId索引
redis.call('SADD', NameKey, AccId)

-- 设置UserId->AccId索引
redis.call('HDEL', UserIdKey, "AccId")
redis.call('HMSET', UserIdKey, "AccId", AccId, "Name", Name)

-- 设置AccId->Name
redis.call('HSET', AccIdKey, "Name", Name)

-- 设置Account -> AccId索引
redis.call('DEL', AccountKey)
redis.call('SET', AccountKey, AccId)

return 0
