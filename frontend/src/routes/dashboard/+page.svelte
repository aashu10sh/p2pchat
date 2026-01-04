<script lang="ts">
	import type { PageProps } from './$types';
	let { data }: PageProps = $props();
</script>

<div>
	<h1>Hello {data.user.user_name}</h1>
</div>
