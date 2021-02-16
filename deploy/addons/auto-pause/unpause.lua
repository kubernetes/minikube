local function unpause(txn, addr, port)
    core.Info("hello_world\n")
    if not addr then addr = '127.0.0.1' end
    if not port then port = 5000 end

    -- Set up a request to the service
    local hdrs = {
        [1] = string.format('host: %s:%s', addr, port),
        [2] = 'accept: */*',
        [3] = 'connection: close'
    }

    local req = {
        [1] = string.format('GET /%s HTTP/1.1', tostring(txn.f:src())),
        [2] = table.concat(hdrs, '\r\n'),
        [3] = '\r\n'
    }

    req = table.concat(req,  '\r\n')

    -- Use core.tcp to get an instance of the Socket class
    local socket = core.tcp()
    socket:settimeout(5)

    -- Connect to the service and send the request
    if socket:connect(addr, port) then
        if socket:send(req) then
            -- Skip response headers
            while true do
                local line, _ = socket:receive('*l')

                if not line then break end
                if line == '' then break end
            end

            -- Get response body, if any
            local content = socket:receive('*a')

            -- Check if this request should be allowed
            if content and content == 'allow' then
                txn:set_var('req.blocked', false)
                return
            end
        else
            core.Alert('Could not connect to IP Checker server (send)')
        end

        socket:close()
    else
        core.Alert('Could not connect to IP Checker server (connect)')
    end

    -- The request should be blocked
    txn:set_var('req.blocked', true)
end

core.register_action('unpause', {'tcp-req'}, unpause, 2)

