<script lang="ts">
	import LoginForm from '$lib/components/login-form.svelte';
	import Logo from '$lib/logo.svelte';
	import type { PageProps } from './$types';

	let { form }: PageProps = $props();
	$inspect(form?.error);
</script>

<div class="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-4">
		<a href="/" class="flex items-center gap-2 self-center text-xl font-medium text-white/60">
			<Logo class="size-6 " />
			WatchTower
		</a>
		<LoginForm error={form?.error} />
	</div>
</div>
