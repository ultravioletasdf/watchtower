<script lang="ts">
	let { data } = $props();
</script>

{#await data.streamed.videos}
	Loading...
{:then videos}
	{#if videos.videos}
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each videos.videos as video (video.id)}
				<div class="flex flex-col gap-1">
					<div class="bg-card border-border aspect-video rounded-lg border"></div>
					<div class="font-semibold">
						{video.title}
					</div>
				</div>
			{:else}
				No videos...
			{/each}
		</div>
	{/if}
{/await}
