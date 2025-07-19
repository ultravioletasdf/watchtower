<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import * as Card from "$lib/components/ui/card/index.js";
	import { Label } from "$lib/components/ui/label/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { cn } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";
	import { onMount } from "svelte";
	import { enhance } from "$app/forms";
	import type { Error } from "$lib/types";

	let { class: className, error, ...restProps }: HTMLAttributes<HTMLDivElement> & {error: Error | undefined}= $props();
	const id = $props.id();

	let disabled = $state(true);
	let email: HTMLInputElement|null = $state(null)
	let password: HTMLInputElement|null = $state(null)

	onMount(() => {
		updateButton()
	})
	function updateButton(){
		if (!email || !password) return
		disabled = !email.checkValidity() || !password.checkValidity()
	}
</script>

<div class={cn("flex flex-col gap-6", className)} {...restProps}>
	<Card.Root>
		<Card.Header class="text-center">
			<Card.Title class="text-xl">Welcome back</Card.Title>
			<Card.Description>
				{#if error && error.incorrectField == "" && error.message != ""}
				<div class="text-error">
					{error.message}
				</div>
				{/if}
				Login to continue to WatchTower</Card.Description>
		</Card.Header>
		<Card.Content>
			<form use:enhance method="POST">
				<div class="grid gap-6">
					<div class="grid gap-6">
						<div class="grid gap-1.5">
							<Label for="email-{id}">Email</Label>
							<Input
								id="email-{id}"
								name="email"
								type="email"
								aria-invalid={error?.incorrectField == "email"}
								placeholder="m@example.com"
								required
								oninput={updateButton}
								bind:ref={email}
							/>
							{#if error?.incorrectField == "email"}
							<p class="text-red-500 text-sm">{error.message}</p>
							{/if}
						</div>
						<div class="grid gap-1.5">
							<div class="flex items-center">
								<Label for="password-{id}">Password</Label>
								<a
									href="##"
									class="ml-auto text-sm underline-offset-4 hover:underline"
								>
									Forgot your password?
								</a>
							</div>
							<Input id="password-{id}" name="password" type="password" required minlength={8} maxlength={72} oninput={updateButton} bind:ref={password}/>
							{#if error?.incorrectField == "password"}
							<p class="text-red-500 text-sm">{error.message}</p>
							{/if}
						</div>
						<Button type="submit" class="w-full" {disabled}>Login</Button>
					</div>
					<div class="text-center text-sm">
						Don&apos;t have an account?
						<a href="/sign/up" class="underline underline-offset-4"> Sign up </a>
					</div>
				</div>
			</form>
		</Card.Content>
	</Card.Root>
	<div
		class="text-muted-foreground *:[a]:hover:text-primary *:[a]:underline *:[a]:underline-offset-4 text-balance text-center text-xs"
	>
		By clicking continue, you agree to our <a href="##">Terms of Service</a>
		and <a href="##">Privacy Policy</a>.
	</div>
</div>