<script lang="ts">
	import Dropzone from 'svelte-file-dropzone';
	import { goto } from '$app/navigation';
	import { upload } from '$lib/utils';
	let uploading = $state(false);

	async function handleVideoDrop(e: CustomEvent<any>) {
		uploading = true;
		const id = await upload(e, '/upload');
		goto(`/upload/${id}`);
	}
</script>

<div class="flex items-center justify-center">
	<Dropzone
		on:drop={handleVideoDrop}
		accept="video/*"
		containerClasses="bg-card! border-border! rounded-lg! disabled:opacity/80! min-h-40 justify-center!"
		disabled={uploading}
	>
		{#if uploading}
			Uploading...
		{:else}
			Drag a video here or click to select a file
		{/if}
	</Dropzone>
</div>
