--[[
-- usage: offline.lua keyname accid plat userid gid
-- keyname: 表名，名字一般为 online[_{who}]:accid
-- accid: 账号id
-- plat: 渠道
-- userid: openid
-- gid: gamesvr id
--
-- hashtable online_derrick:accid
-- 情况-0: 参数个数错误，返回 0
]]--

if #ARGV ~= 5 then
    return 0
end

local accIdKey = ARGV[1] .. ":" .. ARGV[2]
local userIdKey = ARGV[1] .. ":" .. ARGV[3] .. ":" .. ARGV[4]
local gid = ARGV[5]

local data = redis.call("hgetall", userIdKey)
if data ~= nil and data[2] == gid then
    return redis.call("del", accIdKey, userIdKey)
end

return 2

