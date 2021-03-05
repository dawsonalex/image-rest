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

<div class="image-template" style="--width: {width};--height: {height}">
    <img alt="altText" data-url="{url}" data-width="{width}" data-height="{height}" />
</div>

<style>
    .image-template {
        background-color: #999999;
        width: var(--width);
        height: var(--height);
    }
</style>