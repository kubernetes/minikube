local function foo(txn)
    core.Info("hello_world\n")
    f = io.open("/tmp/abc.txt", "a")
    f:write("hello_world\n")
    f:close()
end

core.register_action("foo_action", { 'tcp-req' }, foo, 0)

