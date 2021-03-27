<script>
    import {onMount} from 'svelte';
    import LazyImage from './LazyImage.svelte';

    let imageData = [];

    onMount(async () => {
        const res = await fetch(`http://localhost:8080/list`);
        if (res.ok) {
            // TODO: Image data needs to be scaled, or thumbnails sent from backend, since original images are too big.
            imageData = await res.json();
            console.log(imageData);
        } else {
            console.log('SOMETHING WENT WRONG');
        }
    });
</script>

<section id="photos">
    {#each imageData as image}
        <!-- TODO: Load API URLs from some main config, don't hardcode them. -->
        <LazyImage height="{image.height}" width="{image.width}" url="http://localhost:8080/image?name={image.name}" />
    {/each}
</section>

<style>
    #photos {
        /* Prevent vertical gaps */
        line-height: 0;

        -webkit-column-count: 5;
        -webkit-column-gap:   0;
        -moz-column-count:    5;
        -moz-column-gap:      0;
        column-count:         5;
        column-gap:           0;
    }

    /*#photos LazyImage {*/
    /*    !* Just in case there are inline attributes *!*/
    /*    width: 100% !important;*/
    /*    height: auto !important;*/
    /*}*/

    @media (max-width: 1200px) {
        #photos {
            -moz-column-count:    4;
            -webkit-column-count: 4;
            column-count:         4;
        }
    }
    @media (max-width: 1000px) {
        #photos {
            -moz-column-count:    3;
            -webkit-column-count: 3;
            column-count:         3;
        }
    }
    @media (max-width: 800px) {
        #photos {
            -moz-column-count:    2;
            -webkit-column-count: 2;
            column-count:         2;
        }
    }
    @media (max-width: 400px) {
        #photos {
            -moz-column-count:    1;
            -webkit-column-count: 1;
            column-count:         1;
        }
    }
</style>