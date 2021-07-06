import H from 'history'
import React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

import { SymbolUsagePatternFields } from '../../../graphql-operations'

import {
    SymbolUsagePatternExampleLocation,
    SymbolUsagePatternExampleLocationGQLFragment,
} from './SymbolUsagePatternExampleLocation'

const HACK_CSS = (
    <style>
        {
            'td.line { display: none; } .code-excerpt .code { padding-left: 0.25rem !important; } .result-container__header { display: none; } .result-container { border: solid 1px var(--border-color) !important; border-width: 1px !important; margin: 1rem; }'
        }
    </style>
)

export const SymbolUsagePatternGQLFragment = gql`
    fragment SymbolUsagePatternFields on SymbolUsagePattern {
        description

        exampleLocations {
            ...SymbolUsagePatternExampleLocationFields
        }
    }
    ${SymbolUsagePatternExampleLocationGQLFragment}
`

interface Props extends SettingsCascadeProps, ThemeProps, VersionContextProps {
    usagePatterns: SymbolUsagePatternFields[]

    location: H.Location
}

export const SymbolUsagePatternsSection: React.FunctionComponent<Props> = ({ usagePatterns, ...props }) => (
    <>
        {HACK_CSS}
        {usagePatterns.map(({ description, exampleLocations }, index) => (
            <div key={index}>
                {false && (
                    <h4>
                        <Markdown dangerousInnerHTML={renderMarkdown(description)} />
                    </h4>
                )}
                {exampleLocations.map((exampleLocation, index) => (
                    <SymbolUsagePatternExampleLocation key={index} exampleLocation={exampleLocation} {...props} />
                ))}
            </div>
        ))}
    </>
)
