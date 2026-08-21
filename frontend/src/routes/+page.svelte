<script lang="ts">
    import type { RequestLog } from './+page.server';

    // Define a clean Props interface separately to keep the parser happy
    interface Props {
        data: {
            metrics: RequestLog[];
        };
    }

    // Destructure data cleanly using our Props interface
    let { data }: Props = $props();
</script>

<div class="p-8">
    <h1 class="text-2xl font-bold mb-4">Backend Connection Test</h1>
    
    {#if data.metrics.length === 0}
        <p class="text-red-500 font-semibold">No data found, or the Go backend is down.</p>
    {:else}
        <p class="text-green-600 font-semibold mb-4">Successfully fetched {data.metrics.length} records from Go/PostgreSQL!</p>
        
        <pre class="bg-gray-800 text-green-400 p-4 rounded-lg overflow-auto shadow-lg text-sm">
{JSON.stringify(data.metrics, null, 2)}
        </pre>
    {/if}
</div>