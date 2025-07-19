<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import HouseIcon from '@lucide/svelte/icons/house';
	import HeartIcon from '@lucide/svelte/icons/heart';
	import SearchIcon from '@lucide/svelte/icons/search';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';

	// Menu items.
	const items = [
		{
			title: 'Home',
			url: '#',
			icon: HouseIcon
		},
		{
			title: 'Following',
			url: '/following',
			icon: HeartIcon
		}
	];
	const footerItems = [
		{
			title: 'Settings',
			url: '#',
			icon: SettingsIcon
		}
	];
</script>

<Sidebar.Root class="relative h-[calc(100dvh-68px)]">
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each items as item (item.title)}
						<Sidebar.MenuItem>
							<Sidebar.MenuButton>
								{#snippet child({ props })}
									<a href={item.url} {...props}>
										<item.icon />
										<span>{item.title}</span>
									</a>
								{/snippet}
							</Sidebar.MenuButton>
						</Sidebar.MenuItem>
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<Sidebar.Menu>
			{#each footerItems as item (item.title)}
				<Sidebar.MenuItem>
					<Sidebar.MenuButton>
						{#snippet child({ props })}
							<a href={item.url} {...props}>
								<item.icon />
								<span>{item.title}</span>
							</a>
						{/snippet}
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			{/each}
		</Sidebar.Menu>
	</Sidebar.Footer>
</Sidebar.Root>
