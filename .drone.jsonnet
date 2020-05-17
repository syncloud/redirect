local build(arch) = {
    kind: "pipeline",
    name: arch,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
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
        {                                                   name: "build",                                  image: "syncloud/build-deps-" + arch,
            commands: [
                "./build.sh",
            ]
        }
    ]
};

[
    build("amd64")
]
