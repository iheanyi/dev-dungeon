-- Fix corrupted item template IDs in game saves
-- Old bug saved instance IDs (e.g., "item_malloc_123") instead of template IDs ("malloc")
-- This migration extracts the template ID from the instance ID format

-- Create a function to fix template IDs
CREATE OR REPLACE FUNCTION fix_template_id(template_id TEXT) RETURNS TEXT AS $$
DECLARE
    parts TEXT[];
    result TEXT;
BEGIN
    -- If it doesn't start with 'item_', it's already correct
    IF template_id IS NULL OR NOT template_id LIKE 'item_%' THEN
        RETURN template_id;
    END IF;

    -- Split by underscore: item_malloc_123 -> ['item', 'malloc', '123']
    -- or item_vim_blade_456 -> ['item', 'vim', 'blade', '456']
    parts := string_to_array(template_id, '_');

    -- If we have less than 3 parts, return as-is
    IF array_length(parts, 1) < 3 THEN
        RETURN template_id;
    END IF;

    -- Check if the last part is numeric (the instance suffix)
    IF parts[array_length(parts, 1)] ~ '^\d+$' THEN
        -- Remove first element ('item') and last element (numeric suffix)
        -- Rejoin with underscores
        result := array_to_string(parts[2:array_length(parts, 1)-1], '_');
        RETURN result;
    END IF;

    -- If last part isn't numeric, return as-is
    RETURN template_id;
END;
$$ LANGUAGE plpgsql;

-- Update all game saves
UPDATE game_saves
SET save_data = jsonb_set(
    jsonb_set(
        save_data,
        '{player,inventory}',
        COALESCE(
            (SELECT jsonb_agg(
                jsonb_set(item, '{template_id}', to_jsonb(fix_template_id(item->>'template_id')))
            )
            FROM jsonb_array_elements(save_data->'player'->'inventory') AS item),
            '[]'::jsonb
        )
    ),
    '{player,equipment}',
    jsonb_build_object(
        'weapon', fix_template_id(save_data->'player'->'equipment'->>'weapon'),
        'armor', fix_template_id(save_data->'player'->'equipment'->>'armor'),
        'utility1', fix_template_id(save_data->'player'->'equipment'->>'utility1'),
        'utility2', fix_template_id(save_data->'player'->'equipment'->>'utility2')
    )
)
WHERE save_data->'player' IS NOT NULL;

-- Clean up the function
DROP FUNCTION fix_template_id(TEXT);
