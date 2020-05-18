local name = "redirect";

local build(arch) = {
    kind: "pipeline",
    name: name,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "build",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./build.sh",
            ]
        },
        {
            name: "test",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./test.deps.sh",
                "./configure test",
                "./ci/redirectdb create redirect",
                "py.test --cov redirect"
            ]
        },
        {
            name: "deploy",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./ci/deploy"
            ]
        },
        {
            name: "artifact",
            image: "appleboy/drone-scp",
            settings: {
                host: {
                    from_secret: "artifact_host"
                },
                username: "artifact",
                key: {
                    from_secret: "artifact_key"
                },
                timeout: "2m",
                command_timeout: "2m",
                target: "/home/artifact/repo/" + name + "/${DRONE_BUILD_NUMBER}" ,
                source: "artifact/*",
                     strip_components: 1
            },
            when: {
              status: [ "failure", "success" ]
            }
        }
    ]
};

[
    build("amd64")
]
