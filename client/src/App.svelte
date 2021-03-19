<script>
	import ImageGrid from './ImageGrid.svelte';
	import {onMount} from "svelte";

	export let name;

	function upload() {
		console.log('Uploading files');
		let formData = new FormData();
		let fileInput = document.getElementById('image-upload');
		let files = fileInput.files;
		formData.append('image', files);
		fetch('/upload', {method: 'POST', body: formData}).then(response => console.log(response.body));
	}

	onMount(async () => {
		document.getElementById('image-upload').addEventListener('change', upload);
	});
</script>

<main>
	<h1>Hello {name}!</h1>
	<p>Visit the <a href="https://svelte.dev/tutorial">Svelte tutorial</a> to learn how to build Svelte apps.</p>
	<ImageGrid />
	<form id="image-form" action="http://localhost:8080/upload" enctype="multipart/form-data" method="post" novalidate>
		<input id="image-upload" class="upload-button" type="file" accept=".jpeg,.jpg,.png" multiple/>
	</form>
</main>

<style>
	main {
		text-align: center;
		padding: 1em;
		max-width: 240px;
		margin: 0 auto;
	}

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 4em;
		font-weight: 100;
	}

	.upload-button {
		width: 4rem;
		height: 4rem;
		border-radius: 50%;
		background-color: deepskyblue;
		font-size: 3rem;
		line-height: 1rem;

		position: fixed;
		bottom: 2rem;
		right: 2rem;
		cursor: pointer;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
</style>
