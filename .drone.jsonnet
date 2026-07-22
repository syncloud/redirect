local name = "redirect";
local go = "1.25";
local dind = "19.03.8-dind";
local node = "18.12.0";
local playwright = "v1.59.1-jammy";
local platform = "26.04.2";
local docker_image = "syncloud/redirect";
local version = "${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}";
local image_tag = docker_image + ":" + version;

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
            image: "node:" + node,
            commands: [
                "mkdir build",
                "cd www",
                "bash ../ci/npm.sh install",
                "npm run test",
                "npm run lint",
                "npm run build",
                "cp -r dist ../build/www"
            ]
        },
        {
            name: "test backend",
            image: "golang:" + go,
            commands: [
                "./backend/test.sh",
            ]
        },
        {
            name: "build backend",
            image: "golang:" + go,
            commands: [
                "./backend/build.sh ${DRONE_BUILD_NUMBER}",
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
            name: "docker",
            image: "plugins/docker:20.18",
            settings: {
                repo: docker_image,
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                tags: [
                    version,
                    "${DRONE_BRANCH}",
                ],
            },
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "docker latest",
            image: "plugins/docker:20.18",
            settings: {
                repo: docker_image,
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                tags: ["latest"],
            },
            when: {
                event: ["push"],
                branch: ["stable"],
            },
        },
        {
            name: "docker caddy",
            image: "plugins/docker:20.18",
            settings: {
                repo: "syncloud/caddy",
                dockerfile: "docker/caddy/Dockerfile",
                context: "docker/caddy",
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                tags: ["latest"],
            },
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "deploy test",
            image: "debian:bookworm-slim",
            environment: {
                DEPLOY_HOST: "www.syncloud.test",
                DEPLOY_USER: "root",
                DEPLOY_URL: "https://api.syncloud.test",
                DEPLOY_ENV: "integration",
                SYNCLOUD_DOMAIN: "syncloud.test",
                DB_HOST: "mysql",
                access_key_id: { from_secret: "access_key_id" },
                secret_access_key: { from_secret: "secret_access_key" },
                hosted_zone_id: { from_secret: "hosted_zone_id" },
                PAYPAL_URL: "https://api-m.sandbox.paypal.com",
                PAYPAL_PLAN_MONTHLY_ID: "1",
                PAYPAL_PLAN_ANNUAL_ID: "2",
                PAYPAL_CLIENT_ID: "3",
                PAYPAL_SECRET_ID: "4",
                STRIPE_SECRET_KEY: "sk_test_dummy",
                STRIPE_PRICE_MONTHLY_ID: "price_dummy_monthly",
                STRIPE_PRICE_ANNUAL_ID: "price_dummy_annual",
            },
            commands: [
                "./ci/test-init.sh",
                "./ci/test-setup.sh",
                "./ci/deploy-prepare.sh",
                "./ci/deploy-run.sh " + image_tag,
                "./ci/deploy-verify.sh",
            ],
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "test-api",
            image: "python:3.9-slim-bullseye",
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
                "cd integration",
                "py.test -x -vv -s test.py --domain=syncloud.test --device-host=www.syncloud.test --build-number=${DRONE_BUILD_NUMBER}"
            ],
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "test-ui-desktop",
            image: "mcr.microsoft.com/playwright:" + playwright,
            environment: {
                CI: "true",
                PLAYWRIGHT_DOMAIN: "syncloud.test"
            },
            commands: [
                "./ci/ui.sh desktop"
            ]
        },
        {
            name: "test-ui-mobile",
            image: "mcr.microsoft.com/playwright:" + playwright,
            environment: {
                CI: "true",
                PLAYWRIGHT_DOMAIN: "syncloud.test"
            },
            commands: [
                "./ci/ui.sh mobile"
            ]
        },
        {
            name: "deploy uat",
            image: "debian:bookworm-slim",
            environment: {
                DEPLOY_HOST: { from_secret: "uat_deploy_host" },
                DEPLOY_USER: { from_secret: "uat_deploy_user" },
                DEPLOY_KEY: { from_secret: "uat_deploy_key" },
                DEPLOY_URL: { from_secret: "uat_deploy_url" },
                PAYPAL_URL: "https://api-m.sandbox.paypal.com",
                PAYPAL_PLAN_MONTHLY_ID: { from_secret: "uat_paypal_plan_monthly_id" },
                PAYPAL_PLAN_ANNUAL_ID: { from_secret: "uat_paypal_plan_annual_id" },
                PAYPAL_CLIENT_ID: { from_secret: "uat_paypal_client_id" },
                PAYPAL_SECRET_ID: { from_secret: "uat_paypal_secret_id" },
                STRIPE_SECRET_KEY: { from_secret: "uat_stripe_secret_key" },
                STRIPE_PRICE_MONTHLY_ID: { from_secret: "uat_stripe_price_monthly_id" },
                STRIPE_PRICE_ANNUAL_ID: { from_secret: "uat_stripe_price_annual_id" },
            },
            commands: [
                "./ci/deploy-prepare.sh",
                "./ci/deploy-run.sh " + image_tag,
                "./ci/deploy-verify.sh",
                "./ci/grafana-deploy.sh",
            ],
            when: { event: ["push"] },
        },
        {
            name: "deploy prod",
            image: "debian:bookworm-slim",
            environment: {
                DEPLOY_HOST: { from_secret: "prod_deploy_host" },
                DEPLOY_USER: { from_secret: "prod_deploy_user" },
                DEPLOY_KEY: { from_secret: "prod_deploy_key" },
                DEPLOY_URL: { from_secret: "prod_deploy_url" },
                PAYPAL_URL: "https://api-m.paypal.com",
                PAYPAL_PLAN_MONTHLY_ID: { from_secret: "prod_paypal_plan_monthly_id" },
                PAYPAL_PLAN_ANNUAL_ID: { from_secret: "prod_paypal_plan_annual_id" },
                PAYPAL_CLIENT_ID: { from_secret: "prod_paypal_client_id" },
                PAYPAL_SECRET_ID: { from_secret: "prod_paypal_secret_id" },
                STRIPE_SECRET_KEY: { from_secret: "prod_stripe_secret_key" },
                STRIPE_PRICE_MONTHLY_ID: { from_secret: "prod_stripe_price_monthly_id" },
                STRIPE_PRICE_ANNUAL_ID: { from_secret: "prod_stripe_price_annual_id" },
            },
            commands: [
                "./ci/deploy-prepare.sh",
                "./ci/deploy-run.sh " + image_tag,
                "./ci/deploy-verify.sh",
            ],
            when: { event: ["push"], branch: ["stable"] },
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
            name: "www.syncloud.test",
            image: "syncloud/bootstrap-bookworm-amd64:" + platform,
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
            image: "docker:" + dind,
            environment: {
                DOCKER_USERNAME: {
                    from_secret: "DOCKER_USERNAME"
                },
                DOCKER_PASSWORD: {
                    from_secret: "DOCKER_PASSWORD"
                }
            },
            commands: [
                "cd docker",
                "./push-redirect-test.sh " + arch
            ],
            volumes: [
                {
                    name: "dockersock",
                    path: "/var/run"
                }
            ],
            when: {
                branch: ["stable", "master"]
            }
        },
    ],
    services: [
        {
            name: "docker",
            image: "docker:" + dind,
            privileged: true,
            volumes: [
                {
                    name: "dockersock",
                    path: "/var/run"
                }
            ]
        }
    ],
    volumes: [
        {
            name: "dbus",
            host: {
                path: "/var/run/dbus"
            }
        },
        {
            name: "dockersock",
            temp: {}
        }
    ]
}];

build("amd64") +
build_testapi("amd64") +
build_testapi("arm64") +
build_testapi("arm")
