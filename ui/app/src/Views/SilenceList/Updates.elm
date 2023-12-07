module Views.SilenceList.Updates exposing (update)

import Browser.Navigation as Navigation
import Data.GettableSilence exposing (GettableSilence)
import Data.SilenceStatus exposing (State(..))
import Silences.Api as Api
import Utils.Api as ApiData
import Utils.Filter exposing (Filter)
import Utils.Types exposing (ApiData(..))
import Views.FilterBar.Updates as FilterBar
import Views.SilenceList.Types exposing (Model, SilenceListMsg(..), SilenceTab)


update : SilenceListMsg -> Model -> Filter -> String -> String -> ( Model, Cmd SilenceListMsg )
update msg model filter basePath apiUrl =
    case msg of
        SilencesFetch fetchedSilences ->
            ( { model
                | silences =
                    ApiData.map
                        (\silences -> List.map (groupSilencesByState silences) states)
                        fetchedSilences
              }
            , Cmd.none
            )

        FetchSilences ->
            ( { model
                | filterBar = FilterBar.setMatchers filter model.filterBar
                , silences = Loading
                , showConfirmationDialog = Nothing
              }
            , Api.getSilences apiUrl filter SilencesFetch
            )

        ConfirmDestroySilence silence ->
            ( { model | showConfirmationDialog = Just silence.id }
            , Cmd.none
            )

        DestroySilence silence refresh ->
            -- TODO: "Deleted id: ID" growl
            -- TODO: Check why POST isn't there but is accepted
            ( { model | silences = Loading, showConfirmationDialog = Nothing }
            , Cmd.batch
                [ Api.destroy apiUrl silence (always FetchSilences)
                , if refresh then
                    Navigation.pushUrl model.key (basePath ++ "#/silences")

                  else
                    Cmd.none
                ]
            )

        MsgForFilterBar subMsg ->
            let
                ( newFilterBar, shouldFilter, cmd ) =
                    FilterBar.update subMsg model.filterBar

                filterBarCmd =
                    Cmd.map MsgForFilterBar cmd

                newUrl =
                    Utils.Filter.toUrl (basePath ++ "#/silences")
                        (Utils.Filter.withMatchers newFilterBar.matchers filter)

                silencesCmd =
                    if shouldFilter then
                        Cmd.batch
                            [ Navigation.pushUrl model.key newUrl
                            , filterBarCmd
                            ]

                    else
                        filterBarCmd
            in
            ( { model | filterBar = newFilterBar }, silencesCmd )

        SetTab tab ->
            ( { model | tab = tab }, Cmd.none )


groupSilencesByState : List GettableSilence -> State -> SilenceTab
groupSilencesByState silences state =
    let
        silencesInTab =
            filterSilencesByState state silences
    in
    { tab = state
    , silences = silencesInTab
    , count = List.length silencesInTab
    }


states : List State
states =
    [ Active, Pending, Expired ]


filterSilencesByState : State -> List GettableSilence -> List GettableSilence
filterSilencesByState state =
    List.filter (filterSilenceByState state)


filterSilenceByState : State -> GettableSilence -> Bool
filterSilenceByState state silence =
    silence.status.state == state
