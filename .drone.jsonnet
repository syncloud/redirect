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
            image: "golang:1.15.6",
            commands: [
                "cd backend",
                "go test ./... -cover",
                "go build -o ../build/bin/api cmd/api/main.go",
                "go build -o ../build/bin/www cmd/www/main.go"
            ]
        },
        {
            name: "package",
            image: "syncloud/build-deps-" + arch,
            commands: [
                "cp -r bin build",
                "cp -r redirect build",
                "cp requirements.txt build",
                "cp -r config build",
                "cp -r db build",
                "cp -r emails build",
                "cp redirect_*.wsgi build",
                "mkdir artifact",
                "tar czf artifact/redirect-${DRONE_BUILD_NUMBER}.tar.gz -C build ."
            ]
        },
        {
            name: "test",
            image: "python:3.9-buster",
            commands: [
                "apt-get update -qq",
                "apt-get install -y -qq mysql-client libmysqlclient-dev",
                "pip install -r requirements.txt",
                "pip install -r dev_requirements.txt",
                "adduser --disabled-password --gecos \"\" test",
                "mkdir mail.root",
                "chown test. mail.root",
                "./ci/recreatedb",
                "py.test --cov redirect"
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
	            "pip install -r dev_requirements.txt",
                "cd integration",
                "py.test -x -vv -s verify.py --domain=syncloud.test --device-host=www.syncloud.test --build-number=${DRONE_BUILD_NUMBER}"
            ]
        },
        {
            name: "test-ui-desktop",
            image: "python:3.9-buster",
            commands: [
              "apt-get update && apt-get install -y sshpass openssh-client default-mysql-client",
              "pip install -r dev_requirements.txt",
              "cd integration",
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
              "pip install -r dev_requirements.txt",
              "cd integration",
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
};

[
    build("amd64")
]
