port module Game exposing (..)

import Array exposing (Array)
import Browser
import Dict exposing (Dict)
import Html exposing (Html, button, div, table, td, text, tr)
import Html.Attributes exposing (style)
import Html.Events exposing (onClick)
import Json.Decode as Decode exposing (Decoder, decodeString)
import Json.Encode


main : Program () Model Msg
main =
    Browser.element { init = init, update = update, view = view, subscriptions = subscriptions }



-- PORTS


port sendMessage : String -> Cmd msg


port messageReceiver : (String -> msg) -> Sub msg



-- MODEL


type Model
    = InQueue
    | InGame GameState


type alias GameState =
    { inTurn : Bool
    , board : Board
    , chosenPiece : Maybe Coord
    }


type alias Board =
    Array (Array Cell)


type alias Cell =
    Maybe Unit


type alias Unit =
    { unit : ChessPiece
    , color : Player
    }


type Player
    = White
    | Black


type alias Coord =
    { x : Int, y : Int }


type ChessPiece
    = Pawn
    | Rook
    | Knight
    | Bishop
    | Queen
    | King


newBoard : Board
newBoard =
    let
        w x =
            Just <| Unit x White

        b x =
            Just <| Unit x Black

        row1 c =
            Array.fromList [ c Rook, c Knight, c Bishop, c Queen, c King, c Bishop, c Knight, c Rook ]
    in
    [ row1 w
    , Array.repeat 8 (w Pawn)
    ]
        ++ List.repeat 4 (Array.repeat 8 Nothing)
        ++ [ Array.repeat 8 (b Pawn)
           , row1 b
           ]
        |> Array.fromList


init : () -> ( Model, Cmd Msg )
init _ =
    ( InQueue, Cmd.none )



-- UPDATE


type Msg
    = WebsocketEvent (Result Decode.Error WsEvent)
    | Play String String
    | SelectCell Coord


type alias WsEvent =
    { msg : WsMsg
    , startFirst : Maybe Bool
    , opponentMove : Maybe ( String, String )
    , winner : Maybe WsWinner
    }


type WsMsg
    = GameStart
    | PlayerTurn
    | IllegalMove
    | GameEnd


type WsWinner
    = Player
    | Opponent


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        WebsocketEvent event ->
            handleWebsocketEvent event model

        SelectCell coord ->
            handleOnClick model coord

        _ ->
            ( model, Cmd.none )


handleWebsocketEvent : Result Decode.Error WsEvent -> Model -> ( Model, Cmd Msg )
handleWebsocketEvent event model =
    case ( event, model ) of
        ( Err _, _ ) ->
            -- TODO: log or something?
            ( model, Cmd.none )

        ( Ok ev, InQueue ) ->
            handleEventInQueue ev model

        ( Ok ev, InGame state ) ->
            handleEventInGame ev model


handleEventInQueue : WsEvent -> Model -> ( Model, Cmd Msg )
handleEventInQueue { msg, startFirst } model =
    case ( msg, startFirst ) of
        ( GameStart, Just first ) ->
            ( InGame (GameState first newBoard Nothing)
            , Cmd.none
            )

        _ ->
            ( model, Cmd.none )


handleEventInGame : WsEvent -> Model -> ( Model, Cmd Msg )
handleEventInGame { msg, opponentMove, winner } model =
    case msg of
        PlayerTurn ->
            ( model, Cmd.none )

        IllegalMove ->
            ( model, Cmd.none )

        GameEnd ->
            ( model, Cmd.none )

        _ ->
            ( model, Cmd.none )


handleOnClick : Model -> Coord -> ( Model, Cmd Msg )
handleOnClick model coord =
    case model of
        InGame state ->
            let
                { chosenPiece } =
                    state
            in
            case chosenPiece of
                Nothing ->
                    ( InGame { state | chosenPiece = Just coord }, Cmd.none )

                Just chosen ->
                    ( InGame { state | chosenPiece = Nothing }, sendMessage "TODO" )

        _ ->
            ( model, Cmd.none )


wsEventDecoder : Decoder WsEvent
wsEventDecoder =
    Decode.map4 WsEvent
        (Decode.field "message" wsMsgDecoder)
        (Decode.maybe (Decode.field "startFirst" Decode.bool))
        (Decode.maybe (Decode.field "opponentMove" wsMoveDecoder))
        (Decode.maybe (Decode.field "winner" wsWinnerDecoder))


wsMoveDecoder : Decoder ( String, String )
wsMoveDecoder =
    Decode.map2 Tuple.pair
        (Decode.index 0 Decode.string)
        (Decode.index 1 Decode.string)


wsMsgMapping : Dict String WsMsg
wsMsgMapping =
    Dict.fromList
        [ ( "gameStart", GameStart )
        , ( "playerTurn", PlayerTurn )
        , ( "illegalMove", IllegalMove )
        , ( "gameEnded", GameEnd )
        ]


wsMsgDecoder : Decoder WsMsg
wsMsgDecoder =
    enumStringDecoder wsMsgMapping


wsWinnerDecoder : Decoder WsWinner
wsWinnerDecoder =
    enumStringDecoder
        (Dict.fromList
            [ ( "player", Player )
            , ( "opponent", Opponent )
            ]
        )


enumStringDecoder : Dict String v -> Decoder v
enumStringDecoder mapping =
    Decode.string
        |> Decode.andThen
            (\str ->
                case Dict.get str mapping of
                    Just value ->
                        Decode.succeed value

                    Nothing ->
                        Decode.fail <| "Unkown enum member: " ++ str
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    messageReceiver (\x -> WebsocketEvent (decodeString wsEventDecoder x))



-- VIEW


view : Model -> Html Msg
view model =
    case model of
        InQueue ->
            waitingPage

        InGame game ->
            gamePage game


waitingPage : Html Msg
waitingPage =
    text "Waiting for player..."


gamePage : GameState -> Html Msg
gamePage =
    drawBoard


drawBoard : GameState -> Html Msg
drawBoard state =
    text "TODO"


renderCell : Cell -> Coord -> Html Msg
renderCell cell coord =
    case cell of
        Just unit ->
            renderUnit unit coord

        Nothing ->
            td [] []


renderUnit : Unit -> Coord -> Html Msg
renderUnit { unit, color } coord =
    let
        icon =
            case unit of
                Pawn ->
                    "o"

                Rook ->
                    "R"

                Knight ->
                    "K"

                Bishop ->
                    "B"

                Queen ->
                    "Q"

                King ->
                    "k"

        clr =
            case color of
                White ->
                    "red"

                Black ->
                    "black"
    in
    td [ style "color" clr ] [ button [ onClick (SelectCell coord) ] [ text icon ] ]


coordToChessNotation : Coord -> String
coordToChessNotation { x, y } =
    let
        x1 =
            Char.toCode 'a' + x |> Char.fromCode

        y1 =
            Char.toCode 'a' + y |> Char.fromCode
    in
    String.fromList [ x1, y1 ]
