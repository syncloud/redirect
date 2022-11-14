local name = "redirect";
local go = "1.18.2-buster";

local build(arch) = [{
    kind: "pipeline",
    name: name + "-" + arch,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "build web",
            image: "node:16.1.0",
            commands: [
                "mkdir build",
                "cd www",
                "npm install",
                "npm run test",
                "npm run lint",
                "npm run build",
                "cp -r dist ../build/www"
            ]
        },
        {
            name: "build backend",
            image: "golang:" + go,
            commands: [
                "cd backend",
                "go test ./... -cover",
                "go build -ldflags '-linkmode external -extldflags -static' -o ../build/bin/api ./cmd/api",
                "go build -ldflags '-linkmode external -extldflags -static' -o ../build/bin/www ./cmd/www",
                "go build -ldflags '-linkmode external -extldflags -static' -o ../build/bin/notification ./cmd/cli/notification"
            ]
        },
        {
            name: "package",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "cp -r bin build",
                "cp -r config build",
                "cp -r db build",
                "cp -r emails build",
                "mkdir artifact",
                "tar czf artifact/redirect-${DRONE_BUILD_NUMBER}.tar.gz -C build ."
            ]
        },
        {
            name: "test-integration",
            image: "python:3.9-buster",
            environment: {
                access_key_id: {
                  from_secret: "access_key_id"
                },
                secret_access_key: {
                  from_secret: "secret_access_key"
                },
                hosted_zone_id: {
                  from_secret: "hosted_zone_id"
                },
            },
            commands: [
                "apt-get update && apt-get install -y sshpass openssh-client default-mysql-client",
	            "pip install -r integration/requirements.txt",
	            "./ci/recreatedb",
                "cd integration",
                "py.test -x -vv -s verify.py --domain=syncloud.test --device-host=www.syncloud.test --build-number=${DRONE_BUILD_NUMBER}"
            ]
        },
        {
            name: "test-ui-desktop",
            image: "python:3.9-buster",
            commands: [
              "apt-get update && apt-get install -y sshpass openssh-client default-mysql-client",
              "cd integration",
              "pip install -r requirements.txt",
              "py.test -x -s test-ui.py --ui-mode=desktop --domain=syncloud.test --device-host=www.syncloud.test ",
            ],
            volumes: [{
                name: "shm",
                path: "/dev/shm"
            }]
        },
        {
            name: "test-ui-mobile",
            image: "python:3.9-buster",
            commands: [
              "apt-get update && apt-get install -y sshpass openssh-client default-mysql-client",
              "cd integration",
              "pip install -r requirements.txt",
              "py.test -x -s test-ui.py --ui-mode=mobile --domain=syncloud.test --device-host=www.syncloud.test ",
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
        },
        {
            name: "selenium",
            image: "selenium/standalone-firefox:4.0.0-beta-3-prerelease-20210402",
            volumes: [{
                name: "shm",
                path: "/dev/shm"
            }]
        },
        {
            name: "www.syncloud.test",
            image: "syncloud/platform-jessie-amd64",
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        }
    ],
    volumes: [
        {
         name: "shm",
         temp: {}
        }
    ]
}];

local build_testapi(arch) = [{
    kind: "pipeline",
    name: name + "-testapi-" + arch,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "build test api",
            image: "golang:" + go,
            commands: [
                "cd backend",
                "go build -ldflags '-linkmode external -extldflags -static' -o ../docker/build/testapi ./cmd/testapi",
            ]
        },
        {
            name: "push redirect-test",
            image: "debian:buster-slim",
            environment: {
                DOCKER_USERNAME: {
                    from_secret: "DOCKER_USERNAME"
                },
                DOCKER_PASSWORD: {
                    from_secret: "DOCKER_PASSWORD"
                }
            },
            commands: [
                "./docker/push-redirect-test.sh " + arch
            ],
            privileged: true,
            network_mode: "host",
            volumes: [
                {
                    name: "docker",
                    path: "/usr/bin/docker"
                },
                {
                    name: "docker.sock",
                    path: "/var/run/docker.sock"
                }
            ]
        },
    ],
    volumes: [
        {
            name: "dbus",
            host: {
                path: "/var/run/dbus"
            }
        },
        {
            name: "docker",
            host: {
                path: "/usr/bin/docker"
            }
        },
        {
            name: "docker.sock",
            host: {
                path: "/var/run/docker.sock"
            }
        }
    ],
    when: {
        branch: ["stable", "master"]
    }
}];

build("amd64") +
build_testapi("amd64") +
build_testapi("arm64") +
build_testapi("arm")
