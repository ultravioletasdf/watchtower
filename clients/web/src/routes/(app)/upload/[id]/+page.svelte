<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import Image from '@lucide/svelte/icons/image';
	import Dropzone from 'svelte-file-dropzone';
	import type { PageData } from './$types';
	import { persistedState } from 'svelte-persisted-state';

	let { data } = $props();
	type Draft = {
		title: string;
		description: string;
	};
	let draft = persistedState<Draft>(
		'draft-' + data.id,
		{ title: '', description: '' },
		{ syncTabs: true }
	);
</script>

<div class="flex h-[calc(100dvh-68px-64px)] items-center justify-center">
	<form class="flex w-1/2 flex-col gap-6">
		<h3 class="text-xl font-bold">Video Details</h3>
		<div class="flex flex-col gap-4">
			<div class="w-fulls flex flex-col gap-1.5">
				<Label for="title" class="gap-0">Title<span class="text-rose-400">*</span></Label>
				<div class="flex w-full items-center gap-2">
					<Input
						type="text"
						id="title"
						placeholder="Enter a video title"
						bind:value={draft.current.title}
					/>
					<span class="text-sm opacity-60">{draft.current.title.length}/100</span>
				</div>
			</div>
			<div class="grid w-full gap-1.5">
				<Label for="message">Description</Label>
				<Textarea
					placeholder="Describe your video"
					id="message"
					bind:value={draft.current.description}
				/>
			</div>
			<div class="5 grid w-full gap-1">
				<Label>Thumbnail</Label>
				<Dropzone
					containerClasses="bg-background! border-border! items-start! p-0! rounded-lg! hover:bg-card!  transition!"
				>
					<div class="flex w-full items-center gap-4">
						<div class="bg-card flex h-[180px] w-[320px] items-center justify-center rounded-l-lg">
							<Image />
						</div>
						<div class="flex flex-grow justify-center">Choose a thumbnail</div>
					</div>
				</Dropzone>
			</div>
			<Button>Submit</Button>
		</div>
	</form>
</div>
