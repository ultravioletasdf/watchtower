<script lang="ts">
	import Search from '@lucide/svelte/icons/search';
	import { Input } from './ui/input/index';
	import Button from './ui/button/button.svelte';
	import * as Avatar from './ui/avatar/index';
	import * as DropDownMenu from './ui/dropdown-menu/index';
	import type { User } from '$lib/types';
	import { DropdownMenu } from 'bits-ui';
	import LogOut from '@lucide/svelte/icons/log-out';
	import UserIcon from '@lucide/svelte/icons/circle-user-round';
	import UploadIcon from '@lucide/svelte/icons/upload';
	let { user }: { user: Promise<User> | undefined } = $props();
	function getInitial(username: string | undefined) {
		if (username && username.length > 0) {
			return username[0].toUpperCase();
		}
		return '';
	}
</script>

<div class="bg-sidebar flex items-center justify-between px-8 py-4">
	<div class="flex items-center gap-2 font-bold">
		<img src="/favicon.svg" class="size-6" alt="logo" />
		WatchTower
	</div>
	<form class="flex w-1/4 items-center gap-2" method="GET" action="/search">
		<Input placeholder="Search" class="flex-grow">o</Input>
		<Button variant="secondary" size="icon">
			<Search />
		</Button>
	</form>
	<DropDownMenu.Root>
		<DropDownMenu.Trigger>
			<Avatar.Root>
				<Avatar.Image src="https://github.com/shadcaduouawhdonxn.png" alt="@shadcn" />
				<Avatar.Fallback>
					{#await user}
						...
					{:then user}
						{getInitial(user?.username)}
					{/await}
				</Avatar.Fallback>
			</Avatar.Root>
		</DropDownMenu.Trigger>
		<DropDownMenu.Content align="end">
			<DropdownMenu.Group>
				<a href="/sign/out" data-sveltekit-reload>
					<DropDownMenu.Item>
						<LogOut />
						Sign Out
					</DropDownMenu.Item>
				</a>
				<a href="/profile">
					<DropDownMenu.Item>
						<UserIcon />
						Profile
					</DropDownMenu.Item>
				</a>
				<a href="/upload">
					<DropDownMenu.Item>
						<UploadIcon />
						Upload A Video
					</DropDownMenu.Item>
				</a>
			</DropdownMenu.Group>
		</DropDownMenu.Content>
	</DropDownMenu.Root>
</div>
