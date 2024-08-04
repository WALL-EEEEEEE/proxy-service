task("run")
    on_run(function()
        import("core.base.task")
        task.run("proto")
        local run_cmd = "go run cmd/http/main.go"
        os.exec(package_cmd)

    end)

task("package")
    on_run(function()
        cprint('${green} packaging ...')
        import("core.base.task")
        task.run("proto")
        local package_cmd = "go build -o /go/bin/gateway cmd/http/main.go"
        os.exec(package_cmd)
    end)

task("proto")
    on_run(function()
        cprint('${green} generating proto ...')
        local proto_gen_cmd = "cd /app/proxy/manager && buf generate proto"
        os.exec(proto_gen_cmd)
    end)