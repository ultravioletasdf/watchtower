<script lang="ts">
	import Dropzone from 'svelte-file-dropzone';
	import { goto } from '$app/navigation';

	let uploading = $state(false);
	async function upload(e: CustomEvent<any>): Promise<string> {
		const files: File[] = e.detail.acceptedFiles;
		const file = files[0];
		if (!file) {
			console.log('No file to upload');
		}
		// Allows unlimited uploads atm - FIX THIS
		let res = await fetch('/upload', { method: 'POST' });
		if (!res.ok) {
			alert('Something went wrong');
			throw "Couldn't get presigned post url";
		}
		const response = (await res.json()) as {
			id: string;
			url: string;
			formData: Record<string, string>;
		};
		const formdata = new FormData();
		for (const [k, v] of Object.entries(response.formData)) {
			formdata.append(k, v);
		}
		formdata.append('file', file);
		uploading = true;
		res = await fetch(response.url, {
			method: 'POST',
			body: formdata
		});
		return response.id;
	}
	async function handleVideoDrop(e: CustomEvent<any>) {
		const id = await upload(e);
		goto(`/upload/${id}`);
	}
</script>

<div class="flex h-[calc(100dvh-68px-64px)] items-center justify-center">
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
