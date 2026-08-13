//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock
package inputport

type AuthRepositoryInputPort interface {
}
