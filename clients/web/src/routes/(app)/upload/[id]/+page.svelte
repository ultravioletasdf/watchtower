<script lang="ts">
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select/index';
	import { Textarea } from '$lib/components/ui/textarea';
	import Image from '@lucide/svelte/icons/image';
	import Dropzone from 'svelte-file-dropzone';
	import { persisted } from 'svelte-persisted-store';
	import type { SubmitFunction } from './$types.js';
	import { goto } from '$app/navigation';
	import { upload } from '$lib/utils.js';

	let { data } = $props();
	let error = $state('');

	type Draft = {
		title: string;
		description: string;
		visibility: string;
		thumbnailId: string;
	};
	let draft = persisted<Draft>(
		'draft-' + data.id,
		{ title: '', description: '', visibility: 'Private', thumbnailId: '' },
		{ syncTabs: true }
	);
	draft.subscribe(() => {
		error = '';
	});
	const handleSubmit: SubmitFunction = () => {
		return async ({ result }) => {
			console.log(result);
			switch (result.type) {
				case 'failure': {
					error = result.data?.details ?? 'There was an unknown error.';
					break;
				}
				case 'redirect': {
					goto(result.location);
					break;
				}
				case 'error': {
					error = 'There was an unknown error.';
					break;
				}
				case 'success': {
					goto('/profile');
					break;
				}
			}
		};
	};
	let uploadingThumbnail = $state(false);
	async function handleThumbnailDrop(e: CustomEvent<unknown>) {
		uploadingThumbnail = true;
		$draft.thumbnailId = await upload(e, '/upload/' + data.id + '/thumbnail');
		uploadingThumbnail = false;
	}
	let thumbnailStyle = $derived(
		$draft.thumbnailId != '' ? `background-image: url('/thumbnails/${$draft.thumbnailId}')` : ''
	);
	$inspect($draft);
</script>

<div class="flex h-[calc(100dvh-68px-64px)] items-center justify-center">
	<form class="flex w-1/2 flex-col gap-6" method="POST" use:enhance={handleSubmit}>
		<h3 class="text-xl font-bold">Video Details</h3>
		<div class="flex flex-col gap-4">
			<div class="w-fulls flex flex-col gap-1.5">
				<Label for="title" class="gap-0">Title<span class="text-rose-400">*</span></Label>
				<div class="flex w-full items-center gap-2">
					<Input
						type="text"
						id="title"
						name="title"
						placeholder="Enter a video title"
						bind:value={$draft.title}
					/>
					<span class="text-sm opacity-60">{$draft.title.length}/100</span>
				</div>
			</div>
			<div class="grid w-full gap-1.5">
				<Label for="description">Description</Label>
				<Textarea
					placeholder="Describe your video"
					id="description"
					name="description"
					bind:value={$draft.description}
				/>
			</div>
			<div class="5 grid w-full gap-1">
				<Label>Thumbnail</Label>
				<Dropzone
					containerClasses="bg-background! border-border! items-start! p-0! rounded-lg! hover:bg-card! transition!"
					on:drop={handleThumbnailDrop}
				>
					<div class="flex w-full items-center gap-4">
						<div
							class="bg-card flex h-[180px] w-[320px] items-center justify-center rounded-l-lg"
							style={thumbnailStyle}
						>
							<Image />
						</div>
						<div class="flex flex-grow justify-center">
							{#if uploadingThumbnail}
								Uploading...
							{:else}
								Choose a thumbnail
							{/if}
						</div>
					</div>
				</Dropzone>
			</div>
			<div class="grid w-full gap-1.5">
				<Label for="visibility">Visibility</Label>
				<Select.Root type="single" name="visibility" bind:value={$draft.visibility}>
					<Select.Trigger class="w-full">{$draft.visibility}</Select.Trigger>
					<Select.Content>
						<Select.Item value="Public" label="Public" />
						<Select.Item value="Unlisted" label="Unlisted" />
						<Select.Item value="Private" label="Private" />
					</Select.Content>
				</Select.Root>
			</div>

			<div class="flex flex-col gap-1">
				<Button type="submit">Submit</Button>
				<div class="w-full text-center text-xs text-red-500">{error}</div>
			</div>
		</div>
	</form>
</div>
