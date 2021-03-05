<script>
    import {onMount} from 'svelte';
    import LazyImage from './LazyImage.svelte';

    let imageData = [];

    onMount(async () => {
        const res = await fetch(`http://localhost:8080/list`);
        if (res.ok) {
            // TODO: Image data needs to be scaled, or thumbnails sent from backend, since oringinal images are too big.
            imageData = await res.json();
            console.log(imageData);
        } else {
            console.log('SOMETHING WENT WRONG');
        }
    });
</script>

{#each imageData as image}
    <!-- TODO: Load API URLs from some main config, don't hardcode them. -->
    <LazyImage height="{image.height}" width="{image.width}" url="http://localhost:8080/image?name={image.url}" />
{/each}