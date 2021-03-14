<script>
    export let url, width, height;

    import {onMount} from 'svelte';

    function isInView(elem) {
        // TODO: Create this func and add to listener to detect when to load image.
    }

    onMount(async () => {
        // TODO: Add lazy loader here.
    });
</script>

<img alt="altText" src="{url}" width="100%" height="100%" />