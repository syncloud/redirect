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
            image: "ruby:2.5",
            commands: [
                "./build.sh ${DRONE_BUILD_NUMBER}",
            ]
        },
        {
            name: "test",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./test.deps.sh",
                "py.test --cov redirect"
            ]
        },
        {
            name: "test-integration",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./deploy.deps.sh",
                "cd artifact",
                "../ci/deploy ${DRONE_BUILD_NUMBER} integration syncloud.test",
                "cd ../integration",
                "py.test -x -s verify.py --domain=syncloud.test",
                "xvfb-run -l --server-args='-screen 0, 1024x4096x24' py.test -x -s test-ui.py --ui-mode=desktop --domain=syncloud.test",
                "xvfb-run -l --server-args='-screen 0, 1024x4096x24' py.test -x -s test-ui.py --ui-mode=mobile --domain=syncloud.test",
            ],
            volumes: [{
              name: "shm",
              path: "/dev/shm"
            }]
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
    ],
    services: [
        {
            name: "statsd",
            image: "statsd/statsd"
        },
        {
            name: "mail",
            image: "mailhog/mailhog:v1.0.0",
            environment: {
                MH_HOSTNAME: "syncloud.test"
            }
        },
        {
            name: "mysql",
            image: "mysql:5.7.30",
            environment: {
                MYSQL_ROOT_PASSWORD: "root"
            }
        }
    ],
    volumes: [
        {
         name: "shm",
         temp: {}
        }
    ]
};

[
    build("amd64")
]
