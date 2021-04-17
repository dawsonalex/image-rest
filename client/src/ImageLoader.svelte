<script>
    export let src;
    export let alt;
    export let width, height;

    import IntersectionObserver from './IntersectionObserver.svelte';
    import Image from './Image.svelte';

</script>

<IntersectionObserver once={true} let:intersecting={intersecting}>
    {#if intersecting}
        <Image {alt} {src} {width} {height} />
    {/if}
</IntersectionObserver>