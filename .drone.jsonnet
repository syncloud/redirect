local build(arch) = {
    kind: "pipeline",
    name: arch,

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "build",
            image: "syncloud/build-deps-" + arch,
            commands: [
	        "./test.deps.sh",
                "./configure test",
                "./ci/redirectdb create redirect",
                "py.test --cov redirect"
            ]
        }
    ]
};

[
    build("amd64")
]
