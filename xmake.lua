target("manager")
    on_run(function()
        import("core.base.task")
        task.run("proto")
        local run_cmd = "go run manager/cmd/server/main.go"
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
        os.cd("manager")
        os.exec("buf generate proto")
    end)