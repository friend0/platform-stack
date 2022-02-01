#!/usr/bin/env bats

@test "SECRETS: Fetch fails when no valid secret ID manifest configured " {
	run stack secrets fetch
	[ "$status" -eq 1 ]
}

@test "SECRETS: Fetch succeeds when valid secret ID manifest with defaults " {
	cd examples/basic
	run stack secrets fetch
	[ "$status" -eq 0 ]

	run bash -c "cat deployments/secrets-local.json | grep no-secret"
	[ "$status" -eq 0 ]
	cd -
}

@test "SECRETS: Fetch succeeds when valid secret ID manifest with explicit parameters " {
	cd examples/basic
	run stack secrets fetch -e local -p utmgsmdev -i deployments
	[ "$status" -eq 0 ]

	run bash -c "cat deployments/secrets-local.json | grep no-secret"
	[ "$status" -eq 0 ]
	cd -
}

@test "SECRETS: Fetch fails when valid secret ID manifest but invalid target environment " {
	cd examples/basic
	run stack secrets fetch -e ci -p utmgsmdev -i deployments
	[ "$status" -eq 1 ]
	cd -
}
