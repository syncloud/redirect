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
                "./build.sh ${DRONE_BUILD_NUMBER}",
            ]
        },
        {
            name: "test",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./test.deps.sh",
                "cp -rf config/test/* .",
                "./ci/redirectdb create redirect",
                "py.test --cov redirect"
            ]
        },
        {
            name: "test-integration",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "./deploy.deps.sh",
                "cd artifact",
                "../ci/deploy ${DRONE_BUILD_NUMBER} uat syncloud.test",
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
     services: [{
        name: "statsd",
        image: "statsd/statsd"
     }],
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
