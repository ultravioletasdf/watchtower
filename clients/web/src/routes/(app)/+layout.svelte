<script lang="ts">
	import AppNav from '$lib/components/app-nav.svelte';
	import AppSidebar from '$lib/components/app-sidebar.svelte';

	let { children, data } = $props();

	import * as Sidebar from '$lib/components/ui/sidebar';
</script>

<div class="flex flex-col">
	<AppNav user={data.streamed?.user} />
	<Sidebar.Provider class="flex-grow">
		<AppSidebar />
		<main class="m-8 flex-grow">
			{@render children?.()}
		</main>
	</Sidebar.Provider>
</div>
