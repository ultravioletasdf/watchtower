<script lang="ts">
	import Toggle from "$lib/Toggle.svelte";
	import type { LayoutProps, PageProps } from "./$types";
	let {data}: LayoutProps = $props()
</script>
<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://svelte.dev/docs/kit">svelte.dev/docs/kit</a> to read the documentation</p>
<Toggle/>
{#await data.streamed?.user}
Loading...
{:then user} 
	{JSON.stringify(user)}
{:catch e}
	Error fetching user: 
	{JSON.stringify(e)}
{/await}